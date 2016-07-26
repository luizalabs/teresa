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
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/luizalabs/paas/api/k8s"
	"github.com/luizalabs/paas/api/models"
	"github.com/luizalabs/paas/api/models/storage"
	"github.com/luizalabs/paas/api/restapi/operations/deployments"
	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/apis/extensions"
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
	sessionIdleInterval    = 1 * time.Second
	builderPodTickDuration = 1 * time.Second
	builderPodWaitDuration = 3 * time.Minute
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

	var err error
	k8sClient, err = unversioned.New(config)
	if err != nil {
		log.Panicf("Erro trying to create a kubernetes client. Error: %s", err.Error())
	}
}

// slugify the input text
// copy from the package github.com/mrvdot/golang-utils/
func slugify(str string) (slug string) {
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

type deployParams struct {
	id         string
	app        string
	team       string
	namespace  string
	storageIn  string
	storageOut string
	slugPath   string
}

// newDeployParams return a struct with the parameters to use for deploy
func newDeployParams(app *storage.Application) *deployParams {
	d := deployParams{
		id:   uuid.New()[:8],
		app:  slugify(app.Name),
		team: slugify(app.Team.Name),
	}
	d.namespace = fmt.Sprintf("%s--%s", d.team, d.app)
	d.storageIn = fmt.Sprintf("deploys/%s/%s/%s/in/app.tar.gz", d.team, d.app, d.id)
	d.storageOut = fmt.Sprintf("deploys/%s/%s/%s/out", d.team, d.app, d.id)
	d.slugPath = fmt.Sprintf("%s/slug.tgz", d.storageOut)
	return &d
}

// uploadArchiveToStorage uploads a file (AppTarball) to storage (AWS S3)
func uploadArchiveToStorage(path *string, file *runtime.File) error {
	log.Printf("starting upload to storage [%s]\n", *path)
	po := &s3.PutObjectInput{
		Bucket: aws.String(storageBucket),
		Body:   file.Data,
		Key:    path,
	}
	defer file.Data.Close()
	if _, err := s3svc.PutObject(po); err != nil {
		log.Printf("error uploading the app tarball to storage, Err: %s\n", err.Error())
		return err
	}
	return nil
}

// deleteArchiveOnStorage deletes a file (AppTarball) from storage (AWS S3)
func deleteArchiveOnStorage(path *string) error {
	log.Printf("deleting archive from storage [%s]\n", *path)
	d := &s3.DeleteObjectInput{
		Bucket: aws.String(storageBucket),
		Key:    path,
	}
	if _, err := s3svc.DeleteObject(d); err != nil {
		log.Printf("error deleting the app tarball from storage, Err: %s\n", err.Error())
		return err
	}
	return nil
}

// buildAppSlug starts a POD who builds the final slug from the AppTarball.
func buildAppSlug(p *deployParams) error {
	buildName := fmt.Sprintf("build--%s--%s", p.app, p.id)
	log.Printf("building the app... builder POD name [%s/%s]", p.namespace, buildName)

	bp := k8s.BuildSlugbuilderPod(false, buildName, p.namespace, p.storageIn, p.storageOut, "")
	builder, err := k8sClient.Pods(p.namespace).Create(bp)
	if err != nil {
		log.Printf("error creating the builder pod for the app. Err: %s\n", err.Error())
		return err
	}

	// wainting buider start
	if err = k8s.WaitForPod(k8sClient, builder.Namespace, builder.Name, sessionIdleInterval, builderPodTickDuration, builderPodWaitDuration); err != nil {
		log.Printf("error when waiting the start of the builder POD. Err: %s\n", err.Error())
		return err
	}
	// waiting builder end
	if err = k8s.WaitForPodEnd(k8sClient, builder.Namespace, builder.Name, builderPodTickDuration, builderPodWaitDuration); err != nil {
		log.Printf("error when waiting the end of the builder POD. Err: %s\n", err.Error())
		return err
	}
	// check the builder exit code
	builder, err = k8sClient.Pods(builder.Namespace).Get(builder.Name)
	if err != nil {
		log.Printf("error trying to discover the builder exit code. Err: %s\n", err.Error())
		return err
	}
	for _, containerStatus := range builder.Status.ContainerStatuses {
		state := containerStatus.State.Terminated
		if state.ExitCode != 0 {
			log.Printf("build pod exited with code %d, stopping deploy.\n", state.ExitCode)
			return err
		}
	}
	// deleting slugbuilder pod from k8s
	if err := k8sClient.Pods(builder.Namespace).Delete(builder.Name, nil); err != nil {
		log.Printf("error trying to delete the builder pod. Err: %s\n", err.Error())
		return err
	}
	return nil
}

// createSlugRunnerDeploy creates a "deis slugRunner" deploy on k8s
func createSlugRunnerDeploy(p *deployParams, a *storage.Application) (deploy *extensions.Deployment, err error) {
	log.Printf("creating k8s deploy [%s/%s]\n", p.namespace, p.app)
	// check for env vars
	env := make(map[string]string)
	for _, e := range a.EnvVars {
		env[e.Key] = e.Value
	}
	d := k8s.BuildSlugRunnerDeployment(p.app, p.namespace, 1, 1, int(a.Scale), p.app, p.slugPath, env)
	// deployment change-cause
	d.Annotations = map[string]string{
		"kubernetes.io/change-cause": fmt.Sprintf("deployUUID:%s", p.id),
	}
	deploy, err = k8sClient.Deployments(p.namespace).Create(d)
	if err != nil {
		log.Printf("error creating deployment. Err: %s\n", err.Error())
	}
	return
}

// updateSlugRunnerDeploySlug search and update the "SLUG_URL" on k8s deploys EnvVars
// FIXME: this should be transformed to a simple helper function
func updateSlugRunnerDeploySlug(p *deployParams, d *extensions.Deployment) (deploy *extensions.Deployment, err error) {
	// updating slug
	for i, e := range d.Spec.Template.Spec.Containers[0].Env {
		if e.Name == "SLUG_URL" {
			e.Value = p.slugPath
			d.Spec.Template.Spec.Containers[0].Env[i] = e
			break
		}
	}
	cause := fmt.Sprintf("deployUUID:%s", p.id)
	deploy, err = updateDeploy(d, cause)
	return
}

func updateDeploy(d *extensions.Deployment, changeCause string) (deploy *extensions.Deployment, err error) {
	log.Printf("updating k8s deploy [%s/%s]\n", d.GetNamespace(), d.GetName())
	// deployment change-cause
	d.Annotations = map[string]string{
		"kubernetes.io/change-cause": changeCause,
	}
	deploy, err = k8sClient.Deployments(d.GetNamespace()).Update(d)
	if err != nil {
		log.Printf("error updating deployment. Err: %s\n", err.Error())
	}
	return
}

// deleteDeploy deletes the deploy from k8s
func deleteDeploy(p *deployParams) error {
	log.Printf("deleting k8s deploy [%s/%s]\n", p.namespace, p.app)
	if err := k8sClient.Deployments(p.namespace).Delete(p.app, nil); err != nil {
		log.Printf("error deleting deployment. Err: %s\n", err.Error())
		return err
	}
	return nil
}

// FIXME: change the model of parameters for getDeploy
// getDeploy gets the deploy from k8s
func getDeploy(p *deployParams) (deploy *extensions.Deployment, err error) {
	log.Printf("get k8s deploy [%s/%s]\n", p.namespace, p.app)
	deploy, err = k8sClient.Deployments(p.namespace).Get(p.app)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		log.Printf("error when checking if deploy exists. Err: %s", err)
	}
	return
}

// createServiceAndGetLBHostName creates the k8s service and wait the exposition of the LoadBalancer... after this, return the loadbalancer
func createServiceAndGetLBHostName(p *deployParams) (lb string, err error) {
	log.Printf("creating service [%s/%s]\n", p.namespace, p.app)
	// create service
	s := k8s.BuildSlugRunnerLBService(p.app, p.namespace, p.app)
	if _, err = k8sClient.Services(p.namespace).Create(s); err != nil {
		log.Printf("error creating the LB for the deployment. Err: %s\n", err.Error())
		return
	}
	// wait for lb to return
	log.Println("waiting for LB hostname")
	err = wait.PollImmediate(3*time.Second, 1*time.Minute, func() (bool, error) {
		log.Println("still waiting LB hostname...")
		var cErr error
		if s, cErr = k8sClient.Services(p.namespace).Get(p.app); cErr != nil {
			return false, cErr
		}
		if len(s.Status.LoadBalancer.Ingress) == 0 {
			return false, nil
		}
		if s.Status.LoadBalancer.Ingress[0].Hostname == "" {
			lb = s.Status.LoadBalancer.Ingress[0].Hostname
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		log.Printf("error getting the hostname of the LB service. Err: %s\n", err.Error())
		return
	}
	log.Printf("service created and LB hostname found [%s]\n", lb)
	return
}

// responder saves informations about the deploy to DB and returns the middleware.Responder object
func responder(p *deployParams, appID uint, description *string, errorDescription string) (resp middleware.Responder) {
	// saving deployment to db...
	sd := storage.Deployment{
		UUID:  p.id,
		AppID: appID,
	}
	if description != nil {
		sd.Description = *description
	}
	if errorDescription != "" {
		log.Printf("deploy finished with error. %s\n", errorDescription)
		sd.Error = errorDescription
		resp = deployments.NewCreateDeploymentDefault(500)
	} else {
		log.Println("deploy finished with success")
		r := deployments.NewCreateDeploymentOK()
		payload := models.Deployment{
			When: strfmt.NewDateTime(),
		}
		r.SetPayload(&payload)
		resp = r
	}
	// save deploy to DB
	storage.DB.Save(&sd)
	return
}

// CreateDeploymentHandler handler triggered when a deploy url is requested
func CreateDeploymentHandler(params deployments.CreateDeploymentParams, principal interface{}) middleware.Responder {
	appID := uint(params.AppID)
	// get app info from DB
	sa := storage.Application{}
	sa.ID = appID
	if storage.DB.Where(&storage.Application{TeamID: uint(params.TeamID)}).Preload("Team").Preload("EnvVars").Preload("Deployments").First(&sa).RecordNotFound() {
		log.Println("app info not found")
		return deployments.NewCreateDeploymentUnauthorized()
	}
	// creating deploy params obj
	x := newDeployParams(&sa)
	log.Printf("starting deploy proccess [%s/%s/%s]\n", x.team, x.app, x.id)
	// upload file
	if err := uploadArchiveToStorage(&x.storageIn, &params.AppTarball); err != nil {
		return responder(x, appID, params.Description, "uploading app tarball")
	}
	// build app
	if err := buildAppSlug(x); err != nil {
		deleteArchiveOnStorage(&x.storageIn)
		return responder(x, appID, params.Description, "building app")
	}

	// creating deploy
	deploy, err := getDeploy(x)
	if err != nil {
		deleteArchiveOnStorage(&x.storageIn)
		return responder(x, appID, params.Description, "creating deploy")
	}
	if deploy == nil { // k8s deploy doesn't exists...
		// creating k8s deployment...
		if _, err = createSlugRunnerDeploy(x, &sa); err != nil {
			deleteArchiveOnStorage(&x.storageIn)
			return responder(x, appID, params.Description, "creating deploy")
		}
		// creating k8s service with LoadBalance...
		lbHostName, err := createServiceAndGetLBHostName(x)
		if err != nil {
			deleteDeploy(x)
			deleteArchiveOnStorage(&x.storageIn)
			return responder(x, appID, params.Description, "creating service")
		}
		// save address fo the LB to db...
		saa := storage.AppAddress{
			Address: lbHostName,
			AppID:   appID,
		}
		storage.DB.Create(&saa)
	} else {
		if _, err := updateSlugRunnerDeploySlug(x, deploy); err != nil {
			deleteArchiveOnStorage(&x.storageIn)
			return responder(x, appID, params.Description, "rolling update deploy")
		}
	}
	return responder(x, appID, params.Description, "")
}
