package handlers

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/kelseyhightower/envconfig"
	"github.com/luizalabs/tapi/k8s"
	"github.com/luizalabs/tapi/models/storage"
	"github.com/luizalabs/tapi/restapi/operations/deployments"
	"github.com/pborman/uuid"
	"k8s.io/kubernetes/pkg/api"
	k8s_errors "k8s.io/kubernetes/pkg/api/errors"
	"k8s.io/kubernetes/pkg/apis/extensions"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/util/wait"
)

const (
	waitConditionTickDuration = 3 * time.Second
)

// K8sConfig struct to accommodate the k8s env config
type K8sConfig struct {
	Host     string `required:"true"`
	Username string `required:"true"`
	Password string `required:"true"`
	Insecure bool   `default:"false"`
}

// BuilderConfig struct to accommodate the builder config
type BuilderConfig struct {
	AwsKey     string        `envconfig:"storage_aws_key"`
	AwsSecret  string        `envconfig:"storage_aws_secret"`
	AwsRegion  string        `envconfig:"storage_aws_region"`
	AwsBucket  string        `envconfig:"storage_aws_bucket"`
	PodTimeout time.Duration `envconfig:"wait_pod_timeout" default:"3m"`
	LBTimeout  time.Duration `envconfig:"wait_lb_timeout" default:"5m"`
}

type flushResponseWriter struct {
	f http.Flusher
	w io.Writer
}

func newFlushResponseWriter(w io.Writer) *flushResponseWriter {
	fw := flushResponseWriter{w: w}
	if f, ok := w.(http.Flusher); ok {
		fw.f = f
	}
	return &fw
}
func (fw flushResponseWriter) Write(p []byte) (n int, err error) {
	n, err = fw.w.Write(p)
	if fw.f != nil {
		fw.f.Flush()
	}
	return
}
func (fw flushResponseWriter) Println(a ...interface{}) (n int, err error) {
	n, err = fw.Write([]byte(fmt.Sprintln(a...)))
	return
}
func (fw flushResponseWriter) Printf(format string, a ...interface{}) (n int, err error) {
	n, err = fw.Write([]byte(fmt.Sprintf(format, a...)))
	return
}

var (
	s3svc         *s3.S3
	k8sClient     *unversioned.Client
	builderConfig BuilderConfig
)

func init() {
	// FIXME: this code below isn't in the best place, change this when it's possible

	// load k8s config from env
	var k8sconf K8sConfig
	err := envconfig.Process("teresak8s", &k8sconf)
	if err != nil {
		log.Fatalf("Failed to read k8s configuration from environment: %s", err.Error())
	}
	// kubernetes
	config := &restclient.Config{
		Host:     k8sconf.Host,
		Username: k8sconf.Username,
		Password: k8sconf.Password,
		Insecure: k8sconf.Insecure,
	}
	k8sClient, err = unversioned.New(config)
	if err != nil {
		log.Panicf("Erro trying to create a kubernetes client. Error: %s", err.Error())
	}

	// load builder config from env
	err = envconfig.Process("teresabuilder", &builderConfig)
	// FIXME: uncomment this and delete the follow line when supporting another storage than AWS S3
	// if err != nil {
	if err != nil || (err == nil && (builderConfig.AwsKey == "" || builderConfig.AwsSecret == "" || builderConfig.AwsRegion == "" || builderConfig.AwsBucket == "")) {
		log.Fatalf("Failed to read the builder storage configuration from environment: %s", err.Error())
	}

	// storage
	awsCredentials := credentials.NewStaticCredentials(builderConfig.AwsKey, builderConfig.AwsSecret, "")
	awsConfig := &aws.Config{
		Region:      &builderConfig.AwsRegion,
		Credentials: awsCredentials,
	}
	s3svc = s3.New(session.New(), awsConfig)
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
	app        *storage.Application
	appName    string
	teamName   string
	namespace  string
	storageIn  string
	storageOut string
	slugPath   string
}

// newDeployParams return a struct with the parameters to use for deploy
func newDeployParams(app *storage.Application) *deployParams {
	d := deployParams{
		id:       uuid.New()[:8],
		app:      app,
		appName:  slugify(app.Name),
		teamName: slugify(app.Team.Name),
	}
	d.namespace = getNamespaceName(d.teamName, d.appName)
	d.storageIn = fmt.Sprintf("deploys/%s/%s/%s/in/app.tar.gz", d.teamName, d.appName, d.id)
	d.storageOut = fmt.Sprintf("deploys/%s/%s/%s/out", d.teamName, d.appName, d.id)
	d.slugPath = fmt.Sprintf("%s/slug.tgz", d.storageOut)
	return &d
}

func getNamespaceName(team, app string) string {
	return fmt.Sprintf("%s--%s", team, app)
}

// uploadArchiveToStorage uploads a file (AppTarball) to storage (AWS S3)
func uploadArchiveToStorage(path *string, file *runtime.File) error {
	log.Printf("starting upload to storage [%s]\n", *path)
	po := &s3.PutObjectInput{
		Bucket: &builderConfig.AwsBucket,
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
		Bucket: &builderConfig.AwsBucket,
		Key:    path,
	}
	if _, err := s3svc.DeleteObject(d); err != nil {
		log.Printf("error deleting the app tarball from storage, Err: %s\n", err.Error())
		return err
	}
	return nil
}

// buildAppSlug starts a POD who builds the final slug from the AppTarball.
func buildAppSlug(p *deployParams, fw *flushResponseWriter) error {
	buildName := fmt.Sprintf("build--%s--%s", p.appName, p.id)
	log.Printf("building the app... builder POD name [%s/%s]", p.namespace, buildName)

	env := make(map[string]string)
	for _, e := range p.app.EnvVars {
		env[e.Key] = e.Value
	}

	bp := k8s.BuildSlugbuilderPod(env, buildName, p.namespace, p.storageIn, p.storageOut, "", false)
	builder, err := k8sClient.Pods(p.namespace).Create(bp)
	if err != nil {
		log.Printf("error creating the builder pod for the app. Err: %s\n", err.Error())
		return err
	}

	// wainting buider start
	if err = k8s.WaitForPod(k8sClient, builder.Namespace, builder.Name, waitConditionTickDuration, builderConfig.PodTimeout); err != nil {
		log.Printf("error when waiting the start of the builder POD. Err: %s\n", err.Error())
		return err
	}

	req := k8sClient.Pods(builder.Namespace).GetLogs(builder.Name, &api.PodLogOptions{
		Follow: true,
	})
	rc, err := req.Stream()
	if err != nil {
		log.Printf("error trying to get builder log. Err: %s\n", err.Error())
		return err
	}
	defer rc.Close()

	io.Copy(fw, rc)

	// waiting builder end
	if err = k8s.WaitForPodEnd(k8sClient, builder.Namespace, builder.Name, waitConditionTickDuration, builderConfig.PodTimeout); err != nil {
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
			msg := fmt.Sprintf("build pod exited with code %d, stopping deploy.\n", state.ExitCode)
			log.Printf(msg)
			return errors.New(msg)
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
	log.Printf("creating k8s deploy [%s/%s]\n", p.namespace, p.appName)
	// check for env vars
	env := make(map[string]string)
	for _, e := range a.EnvVars {
		env[e.Key] = e.Value
	}
	d := k8s.BuildSlugRunnerDeployment(p.appName, p.namespace, 1, 1, int(a.Scale), p.appName, p.slugPath, env)
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
	log.Printf("deleting k8s deploy [%s/%s]\n", p.namespace, p.appName)
	if err := k8sClient.Deployments(p.namespace).Delete(p.appName, nil); err != nil {
		log.Printf("error deleting deployment. Err: %s\n", err.Error())
		return err
	}
	return nil
}

// FIXME: change the model of parameters for getDeploy
// getDeploy gets the deploy from k8s
func getDeploy(p *deployParams) (deploy *extensions.Deployment, err error) {
	log.Printf("get k8s deploy [%s/%s]\n", p.namespace, p.appName)
	deploy, err = k8sClient.Deployments(p.namespace).Get(p.appName)
	if err != nil {
		if k8s_errors.IsNotFound(err) {
			return nil, nil
		}
		log.Printf("error when checking if deploy exists. Err: %s", err)
	}
	return
}

// createServiceAndGetLBHostName creates the k8s service and wait the exposition of the LoadBalancer... after this, return the loadbalancer
func createServiceAndGetLBHostName(p *deployParams) (lb string, err error) {
	log.Printf("creating service [%s/%s]\n", p.namespace, p.appName)
	// create service
	s := k8s.BuildSlugRunnerLBService(p.appName, p.namespace, p.appName)
	if _, err = k8sClient.Services(p.namespace).Create(s); err != nil {
		log.Printf("error creating the LB for the deployment. Err: %s\n", err.Error())
		return
	}
	// wait for lb to return
	log.Println("waiting for LB hostname")
	err = wait.PollImmediate(waitConditionTickDuration, builderConfig.LBTimeout, func() (bool, error) {
		log.Println("still waiting LB hostname...")
		var cErr error
		if s, cErr = k8sClient.Services(p.namespace).Get(p.appName); cErr != nil {
			return false, cErr
		}
		if len(s.Status.LoadBalancer.Ingress) == 0 || (len(s.Status.LoadBalancer.Ingress) != 0 && s.Status.LoadBalancer.Ingress[0].Hostname == "") {
			return false, nil
		}
		lb = s.Status.LoadBalancer.Ingress[0].Hostname
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
func responder(p *deployParams, appID uint, description *string, errorDescription string, fw *flushResponseWriter) {
	errorFound := false
	// saving deployment to db...
	sd := storage.Deployment{
		UUID:  p.id,
		AppID: appID,
	}
	if description != nil {
		sd.Description = *description
	}
	if errorDescription != "" { // in case of error
		sd.Error = errorDescription
		errorFound = true
	}
	storage.DB.Save(&sd)

	if errorFound {
		m := fmt.Sprintf("deploy finished with the error: %s\n", errorDescription)
		log.Print(m)
		fw.Println(m)
		return
	}

	log.Println("deploy finished with success")
	// FIXME: change this... we are doing a select to DB when we used this info some seconds ago
	sa := storage.Application{}
	sa.ID = appID
	// FIXME: getting all deployments when we need only the last one :( i didn't found an easy way to change this
	storage.DB.Preload("Team").Preload("Deployments").Preload("Addresses").First(&sa)

	fw.Printf("\nApp:  %s\n", sa.Name)
	fw.Printf("Scale: %d\n", sa.Scale)
	fw.Println("Addresses:")
	for _, ad := range sa.Addresses {
		fw.Printf("  - %s\n", ad.Address)
	}
	sdeploy := sa.Deployments[len(sa.Deployments)-1] // :(
	fw.Printf("Deploy: %s\n", sdeploy.UUID)
}

type teresaResponder struct {
	params           *deployParams
	appID            uint
	description      *string
	errorDescription string
	fw               *flushResponseWriter
}

func (tr *teresaResponder) run(err error, format string, a ...interface{}) {
	d := ""
	if a != nil {
		d = fmt.Sprintf(format, a)
	}
	if err != nil {
		log.Println("responding error")
		s := fmt.Sprintf("%s\n%v", err.Error(), d)
		responder(tr.params, tr.appID, tr.description, s, tr.fw)
		tr.fw.Write([]byte("----------deployment-error----------"))
	} else {
		log.Println("responding success")
		responder(tr.params, tr.appID, tr.description, d, tr.fw)
		tr.fw.Write([]byte("----------deployment-success----------"))
	}
}

// CreateDeploymentHandler handler triggered when a deploy url is requested
var CreateDeploymentHandler deployments.CreateDeploymentHandlerFunc = func(params deployments.CreateDeploymentParams, principal interface{}) middleware.Responder {
	var r middleware.ResponderFunc = func(rw http.ResponseWriter, pr runtime.Producer) {
		appID := uint(params.AppID)
		// get app info from DB
		sa := storage.Application{}
		sa.ID = appID
		// creating a flush response writer to stream data for the request (text/plain)
		fw := newFlushResponseWriter(rw)
		// check if the user is member of the team before start deploy
		if storage.DB.Where(&storage.Application{TeamID: uint(params.TeamID)}).Preload("Team").Preload("EnvVars").Preload("Deployments").First(&sa).RecordNotFound() {
			log.Println("app info not found")
			fw.Write([]byte("App or team invalid... stopping proccess."))
			// fw.Println("App or team invalid... stopping proccess.")
			return
		}
		// creating deploy params obj
		x := newDeployParams(&sa)
		teresaResp := &teresaResponder{
			params:           x,
			appID:            appID,
			description:      params.Description,
			errorDescription: "",
			fw:               fw,
		}
		log.Printf("starting deploy proccess [%s/%s/%s]\n", x.teamName, x.appName, x.id)
		// upload file
		if err := uploadArchiveToStorage(&x.storageIn, &params.AppTarball); err != nil {
			teresaResp.run(err, "uploading app tarball")
			return
		}
		// build app
		fw.Println("Starting the build")
		if err := buildAppSlug(x, fw); err != nil {
			deleteArchiveOnStorage(&x.storageIn)
			teresaResp.run(err, "building app")
			return
		}
		// creating deploy
		deploy, err := getDeploy(x)
		if err != nil {
			deleteArchiveOnStorage(&x.storageIn)
			teresaResp.run(err, "get existant deploy")
			return
		}
		if deploy == nil { // k8s deploy doesn't exists...
			// creating k8s deployment...
			if _, err = createSlugRunnerDeploy(x, &sa); err != nil {
				deleteArchiveOnStorage(&x.storageIn)
				teresaResp.run(err, "creating deploy")
				return
			}
			// creating k8s service with LoadBalance...
			fw.Println("creating the load balancer...")
			lbHostName, err := createServiceAndGetLBHostName(x)
			if err != nil {
				deleteDeploy(x)
				deleteArchiveOnStorage(&x.storageIn)
				teresaResp.run(err, "creating service")
				return
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
				teresaResp.run(err, "rolling update deploy")
				return
			}
		}
		teresaResp.run(nil, "deployment process finished succesfully")
	}
	return r
}
