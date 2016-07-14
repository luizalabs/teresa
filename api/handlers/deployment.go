package handlers

import (
	"fmt"
	"log"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/luizalabs/paas/api/k8s"
	"github.com/luizalabs/paas/api/models"
	"github.com/luizalabs/paas/api/models/storage"
	"github.com/luizalabs/paas/api/restapi/operations/deployments"
	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/wait"
)

// TODO: this should came from env or conf
const (
	awsAccessKeyID         = "AKIAIUARH63XWZUMCFWA"
	awsSecretAccessKey     = "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"
	storageBucket          = "teresa-staging"
	k8sHost                = "https://k8s-staging.a.luizalabs.com"
	k8sUsername            = "admin"
	k8sPassword            = "VOpgP0Ggnty5mLcq"
	k8sInsecure            = true
	sessionIdleInterval    = time.Duration(1) * time.Second
	builderPodTickDuration = time.Duration(1) * time.Second
	builderPodWaitDuration = time.Duration(3) * time.Minute
)

var (
	s3svc     *s3.S3
	k8sClient *unversioned.Client
)

func init() {
	// storage
	awsCredentials := credentials.NewStaticCredentials(awsAccessKeyID, awsSecretAccessKey, "")
	awsConfig := &aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: awsCredentials,
	}
	s3svc = s3.New(session.New(), awsConfig)

	// kubernetes
	config := &restclient.Config{
		Host:     k8sHost,
		Username: k8sUsername,
		Password: k8sPassword,
		Insecure: k8sInsecure,
	}

	k8sClient, _ = unversioned.New(config)
	// if err != nil {
	// 	log.Panicf("Erro trying to create a kubernetes client. Error: %s", err.Error())
	// }
}

// GenerateSlug from the package github.com/mrvdot/golang-utils/
func GenerateSlug(str string) (slug string) {
	return strings.Map(func(r rune) rune {
		switch {
		case r == ' ', r == '-':
			return '-'
		case r == '_', unicode.IsLetter(r), unicode.IsDigit(r):
			return r
		default:
			return -1
		}
	}, strings.ToLower(strings.TrimSpace(str)))
}

// CreateDeploymentHandler creates deploy
func CreateDeploymentHandler(params deployments.CreateDeploymentParams, principal interface{}) middleware.Responder {
	// get app info from DB
	sa := storage.Application{}
	sa.ID = uint(params.AppID)
	if storage.DB.Where(&storage.Application{TeamID: uint(params.TeamID)}).Preload("Team").Preload("EnvVars").Preload("Deployments").First(&sa).RecordNotFound() {
		log.Println("app info not found")
		return deployments.NewCreateDeploymentUnauthorized() // FIXME: check if is this the correct error
	}

	appSlugName := GenerateSlug(sa.Name)
	teamSlugName := GenerateSlug(sa.Team.Name)
	appSlugNamespace := fmt.Sprintf("%s--%s", teamSlugName, appSlugName)
	deployUUID := uuid.New()[:8]
	storageIn := fmt.Sprintf("deploys/%s/%s/%s/in/app.tar.gz", teamSlugName, appSlugName, deployUUID)
	storageOut := fmt.Sprintf("deploys/%s/%s/%s/out", teamSlugName, appSlugName, deployUUID)

	log.Printf("starting deploy [%s/%s/%s]\n", teamSlugName, appSlugName, deployUUID)

	// uploading app tarball to storage
	log.Printf("starting upload to storage [%s]\n", storageIn)
	po := &s3.PutObjectInput{
		Bucket: aws.String(storageBucket),
		Body:   params.AppTarball.Data,
		Key:    &storageIn,
	}
	defer params.AppTarball.Data.Close()
	if _, err := s3svc.PutObject(po); err != nil {
		log.Printf("error uploading the app tarball to storage, Err: %s\n", err.Error())
		return deployments.NewCreateDeploymentDefault(500)
	}

	// builder proccess...
	buildName := fmt.Sprintf("build--%s--%s", appSlugName, deployUUID)
	// FIXME: maybe we should accept extra buildpacks in the future?!?
	log.Printf("building the app; builder POD name [%s/%s]", appSlugNamespace, buildName)
	bp := k8s.BuildSlugbuilderPod(false, buildName, appSlugNamespace, storageIn, storageOut, "")
	builder, err := k8sClient.Pods(appSlugNamespace).Create(bp)
	if err != nil {
		log.Printf("error creating the builder pod for the app. Err: %s\n", err.Error())
		return deployments.NewCreateDeploymentDefault(500)
	}
	// wainting buider start
	if err = k8s.WaitForPod(k8sClient, builder.Namespace, builder.Name, sessionIdleInterval, builderPodTickDuration, builderPodWaitDuration); err != nil {
		log.Printf("error when waiting the start of the builder POD. Err: %s\n", err.Error())
		return deployments.NewCreateDeploymentDefault(500)
	}
	// waiting builder end
	if err = k8s.WaitForPodEnd(k8sClient, builder.Namespace, builder.Name, builderPodTickDuration, builderPodWaitDuration); err != nil {
		log.Printf("error when waiting the end of the builder POD. Err: %s\n", err.Error())
		return deployments.NewCreateDeploymentDefault(500)
	}
	// check the builder exit code
	var p *api.Pod
	if p, err = k8sClient.Pods(builder.Namespace).Get(builder.Name); err != nil {
		log.Printf("error trying to discover the builder exit code. Err: %s\n", err.Error())
		return deployments.NewCreateDeploymentDefault(500)
	}
	for _, containerStatus := range p.Status.ContainerStatuses {
		state := containerStatus.State.Terminated
		if state.ExitCode != 0 {
			log.Printf("build pod exited with code %d, stopping deploy.\n", state.ExitCode)
			return deployments.NewCreateDeploymentDefault(500)
		}
	}

	// deleting slugbuilder pod from k8s
	if err = k8sClient.Pods(builder.Namespace).Delete(builder.Name, nil); err != nil {
		log.Printf("error trying to delete the builder pod. Err: %s\n", err.Error())
		return deployments.NewCreateDeploymentDefault(500)
	}

	// creating k8s deployment...
	appEnv := make(map[string]string)
	for _, e := range sa.EnvVars {
		appEnv[e.Key] = e.Value
	}
	// TODO: maybe it's not necessary to have a name and a selector name
	srd := k8s.BuildSlugRunnerDeployment(appSlugName, appSlugNamespace, 1, 1, int(sa.Scale), appSlugName, fmt.Sprintf("%s/slug.tgz", storageOut), appEnv)
	di := k8sClient.Deployments(appSlugNamespace)
	if _, err = di.Create(srd); err != nil {
		log.Printf("error creating the deployment. Err: %s\n", err.Error())
		return deployments.NewCreateDeploymentDefault(500)
	}

	// creating k8s service with LoadBalance...
	s := k8s.BuildSlugRunnerLBService(appSlugName, appSlugNamespace, appSlugName)
	_, err = k8sClient.Services(appSlugNamespace).Create(s)
	if err != nil {
		log.Printf("error creating the LB for the deployment. Err: %s\n", err.Error())
		return deployments.NewCreateDeploymentDefault(500)
	}

	// waiting lb get the loadbalancer host
	log.Println("getting LB hostname")
	var lbHostName *string
	err = wait.PollImmediate(3*time.Second, 1*time.Minute, func() (bool, error) {
		log.Println("waiting LB hostname...")
		s, sErr := k8sClient.Services(appSlugNamespace).Get(appSlugName)
		if sErr != nil {
			return false, sErr
		}
		if len(s.Status.LoadBalancer.Ingress) == 0 {
			return false, nil
		}
		if s.Status.LoadBalancer.Ingress[0].Hostname == "" {
			return false, nil
		}
		lbHostName = &s.Status.LoadBalancer.Ingress[0].Hostname
		return true, nil
	})

	if err != nil {
		log.Printf("error getting the hostname of the LB service. Err: %s\n", err.Error())
		return deployments.NewCreateDeploymentDefault(500)
	}
	log.Printf("LB hostname is %s", *lbHostName)

	log.Println("deploy finished with success")

	// saving deployment to db...
	d := storage.Deployment{
		UUID:  deployUUID,
		AppID: sa.ID,
	}
	if params.Description != nil {
		d.Description = *params.Description
	}
	storage.DB.Save(&d)

	// save address fo the LB to db...
	saa := storage.AppAddress{
		Address: *lbHostName,
		AppID:   sa.ID,
	}
	storage.DB.Create(&saa)

	// save deploy to db...
	r := deployments.NewCreateDeploymentOK()
	deployment := models.Deployment{
		Description: &appSlugName,
		When:        strfmt.NewDateTime(),
	}
	r.SetPayload(&deployment)
	return r
}
