package handlers

import (
	"fmt"
	"log"
	"time"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/client/restclient"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/luizalabs/paas/api/k8s"
	"github.com/luizalabs/paas/api/models"
	"github.com/luizalabs/paas/api/restapi/operations/deployments"
	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/client/unversioned"
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
	builderPodWaitDuration = time.Duration(1) * time.Minute
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

// CreateDeploymentHandler creates deploy
func CreateDeploymentHandler(params deployments.CreateDeploymentParams, principal interface{}) middleware.Responder {
	// TODO: get app info from DB to this vars
	appName := "wally"
	appDeployVersion := 1
	appDeployID := uuid.New()[:8]
	appNamespace := "default"
	appReplicas := 2
	deployName := fmt.Sprintf("%s-%d-%s", appName, appDeployVersion, appDeployID)

	// TODO: log here about the app pre uploaded
	f := fmt.Sprintf("apps/in/%s/app.tgz", deployName)
	fo := fmt.Sprintf("apps/out/%s", deployName)

	log.Printf("starting deploy '%s'. initializing upload to storage.", f)
	po := &s3.PutObjectInput{
		Bucket: aws.String(storageBucket),
		Body:   params.File.Data,
		Key:    &f,
	}
	r, err := s3svc.PutObject(po)
	if err != nil {
		log.Printf("error when uploading the app tarball to storage, Err %s\n", err.Error())
		// TODO: response with the real error here
	}
	log.Printf("upload done, etag %s", *r.ETag)

	// creating build POD
	var bp *api.Pod
	bp = k8s.BuildSlugbuilderPod(
		true,
		deployName,
		appNamespace,
		f,
		fo, // put path to slug
		"", // buildpacks
	)

	pi := k8sClient.Pods(appNamespace)
	buildPod, err := pi.Create(bp)
	if err != nil {
		log.Fatalf("Error trying to create the building pod (%s)\n", err)
	}
	log.Printf("building the app. podname: %s\n", buildPod.GetName())
	// waiting pod to start
	if err := k8s.WaitForPod(k8sClient, buildPod.Namespace, buildPod.Name, sessionIdleInterval, builderPodTickDuration, builderPodWaitDuration); err != nil {
		log.Printf("watching events for builder pod startup (%s)\n", err)
	}
	// waiting for to end
	if err := k8s.WaitForPodEnd(k8sClient, buildPod.Namespace, buildPod.Name, builderPodTickDuration, builderPodWaitDuration); err != nil {
		log.Printf("error getting builder pod status (%s)\n", err)
	}
	log.Println("checking the pod exit code")

	p, err := k8sClient.Pods(buildPod.Namespace).Get(buildPod.Name)
	if err != nil {
		log.Printf("error getting builder pod status (%s)", err)
	}
	for _, containerStatus := range p.Status.ContainerStatuses {
		state := containerStatus.State.Terminated
		if state.ExitCode != 0 {
			log.Fatalf("build pod exited with code %d, stopping build.", state.ExitCode)
		}
	}
	log.Println("build ok, let's continue the deploy...")

	slug := fmt.Sprintf("%s/slug.tgz", fo)
	log.Printf("slug: %s", slug)

	// creating deployment
	d := k8s.BuildSlugRunnerDeployment(
		appName, appNamespace,
		1, 1, appReplicas, appName,
		slug,
		map[string]string{
			"AWS_ACCESS_KEY": "AKIAJVYPW4MGR2DW6QDA",
			"AWS_SECRET_KEY": "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",
		})
	di := k8sClient.Deployments(appNamespace)
	_, err = di.Create(d)
	if err != nil {
		log.Fatalf("error trying to create a deployment. Error: %s", err)
	}

	// creating service
	s := k8s.BuildSlugRunnerLBService(appName, appNamespace, appName)
	_, err = k8sClient.Services(appNamespace).Create(s)
	if err != nil {
		log.Fatalf("error trying to create a service. Error: %s", err)
	}

	log.Println("everything is ok")

	replicas := int64(2)
	resp := deployments.NewCreateDeploymentOK()
	pa := models.Deployment{
		Description: &params.Description,
		Replicas:    &replicas,
		When:        strfmt.NewDateTime(),
	}
	resp.SetPayload(&pa)
	return resp
}
