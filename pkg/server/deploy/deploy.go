package deploy

import (
	"fmt"
	"io"
	"strings"

	log "github.com/Sirupsen/logrus"
	context "golang.org/x/net/context"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/build"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/exec"
	"github.com/luizalabs/teresa/pkg/server/slug"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"github.com/luizalabs/teresa/pkg/server/storage"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
	"github.com/luizalabs/teresa/pkg/server/uid"
)

const (
	ProcfileReleaseCmd = "release"
	runLabel           = "run"
	internalSvcType    = "ClusterIP"
)

type Operations interface {
	Deploy(ctx context.Context, user *database.User, appName string, tarBall io.ReadSeeker, description string) (io.ReadCloser, <-chan error)
	List(user *database.User, appName string) ([]*ReplicaSetListItem, error)
	Rollback(user *database.User, appName, revision string) error
}

type K8sOperations interface {
	CreateOrUpdateDeploy(deploySpec *spec.Deploy) error
	CreateOrUpdateCronJob(cronJobSpec *spec.CronJob) error
	ExposeDeploy(namespace, name, vHost, svcType, portName string, w io.Writer) error
	ReplicaSetListByLabel(namespace, label, value string) ([]*ReplicaSetListItem, error)
	DeployRollbackToRevision(namespace, name, revision string) error
	CreateOrUpdateConfigMap(namespace, name string, data map[string]string) error
	DeleteConfigMap(namespace, name string) error
	IsNotFound(err error) bool
	ContainerExplicitEnvVars(namespace, deployName, containerName string) ([]*app.EnvVar, error)
	WatchDeploy(namespace, deployName string) error
}

type DeployOperations struct {
	appOps      app.Operations
	execOps     exec.Operations
	buildOps    build.Operations
	fileStorage storage.Storage
	k8s         K8sOperations
	opts        *Options
}

func (ops *DeployOperations) Deploy(ctx context.Context, user *database.User, appName string, tarBall io.ReadSeeker, description string) (io.ReadCloser, <-chan error) {
	errChan := make(chan error, 1)
	a, err := ops.appOps.Get(appName)
	if err != nil {
		errChan <- err
		return nil, errChan
	}

	teamName, err := ops.appOps.TeamName(appName)
	if err != nil {
		errChan <- err
		return nil, errChan
	}
	a.Team = teamName

	if !ops.appOps.HasPermission(user, appName) {
		errChan <- auth.ErrPermissionDenied
		return nil, errChan
	}

	confFiles, err := getDeployConfigFilesFromTarBall(tarBall, a.ProcessType)
	if err != nil {
		errChan <- teresa_errors.New(ErrInvalidTeresaYamlFile, err)
		return nil, errChan
	}

	deployId := uid.New()
	buildIn := fmt.Sprintf("deploys/%s/%s/in", a.Name, deployId)
	buildDest := fmt.Sprintf("deploys/%s/%s/out", appName, deployId)

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		err = ops.buildOps.CreateByOpts(ctx, &build.CreateOptions{
			App:       a,
			BuildName: deployId,
			SlugIn:    buildIn,
			SlugDest:  buildDest,
			TarBall:   tarBall,
			Stream:    w,
		})
		if err != nil {
			errChan <- err
			log.WithError(err).WithField("id", deployId).Errorf("Building app %s", appName)
			return
		}
		slugURL := fmt.Sprintf("%s/slug.tgz", buildDest)
		if app.IsCronJob(a.ProcessType) {
			err = ops.createOrUpdateCronJob(a, confFiles, w, slugURL, description)
		} else {
			err = ops.createOrUpdateDeploy(a, confFiles, w, slugURL, description, deployId)
		}
		if err != nil {
			errChan <- err
			return
		}
		if !app.IsCronJob(a.ProcessType) {
			ops.watchDeploy(appName, deployId, w, errChan)
		}
	}()
	return r, errChan
}

func (ops *DeployOperations) runReleaseCmd(a *app.App, deployId, slugURL string, stream io.Writer) error {
	imgs := &spec.Images{
		SlugRunner: ops.opts.SlugRunnerImage,
		SlugStore:  ops.opts.SlugStoreImage,
	}
	podSpec := spec.NewRunner(
		fmt.Sprintf("release-%s-%s", a.Name, deployId),
		slugURL,
		imgs,
		a,
		ops.fileStorage,
		ops.buildLimits(),
		"start",
		ProcfileReleaseCmd,
	)

	fmt.Fprintln(stream, "Running release command")
	if err := ops.podRun(context.Background(), podSpec, stream); err != nil {
		if err == ErrPodRunFail {
			return ErrReleaseFail
		}
		return err
	}
	return nil
}

func (ops *DeployOperations) buildLimits() *spec.ContainerLimits {
	return &spec.ContainerLimits{
		CPU:    ops.opts.BuildLimitCPU,
		Memory: ops.opts.BuildLimitMemory,
	}
}

func (ops *DeployOperations) createOrUpdateDeploy(a *app.App, confFiles *DeployConfigFiles, w io.Writer, slugURL, description, deployId string) error {
	if releaseCmd := confFiles.Procfile[ProcfileReleaseCmd]; releaseCmd != "" {
		if err := ops.runReleaseCmd(a, deployId, slugURL, w); err != nil {
			log.WithError(err).WithField("id", deployId).Errorf("Running release command %s in app %s", releaseCmd, a.Name)
			return err
		}
	}

	imgs := &spec.Images{
		SlugRunner: ops.opts.SlugRunnerImage,
		SlugStore:  ops.opts.SlugStoreImage,
	}
	if confFiles.NginxConf != "" {
		imgs.Nginx = ops.opts.NginxImage
		data := map[string]string{"nginx.conf": confFiles.NginxConf}
		if err := ops.k8s.CreateOrUpdateConfigMap(a.Name, a.Name, data); err != nil {
			log.WithError(err).Errorf("Creating config to nginx of app %s", a.Name)
			return err
		}
	} else {
		err := ops.k8s.DeleteConfigMap(a.Name, a.Name)
		if err != nil && !ops.k8s.IsNotFound(err) {
			return err
		}
	}

	deploySpec := spec.NewDeploy(
		imgs,
		description,
		slugURL,
		ops.opts.RevisionHistoryLimit,
		a,
		confFiles.TeresaYaml,
		ops.fileStorage,
	)

	if err := ops.k8s.CreateOrUpdateDeploy(deploySpec); err != nil {
		log.WithError(err).Errorf("Creating deploy app %s", a.Name)
		return err
	}

	if err := ops.exposeApp(a, w); err != nil {
		log.WithError(err).Errorf("Exposing service %s", a.Name)
		return err
	}
	fmt.Fprintln(w, fmt.Sprintf("The app %s has been successfully deployed", a.Name))
	return nil
}

func (ops *DeployOperations) createOrUpdateCronJob(a *app.App, confFiles *DeployConfigFiles, w io.Writer, slugURL, description string) error {
	imgs := &spec.Images{
		SlugRunner: ops.opts.SlugRunnerImage,
		SlugStore:  ops.opts.SlugStoreImage,
	}
	if confFiles.TeresaYaml == nil || confFiles.TeresaYaml.Cron == nil {
		return ErrCronScheduleNotFound
	}
	cronSpec := spec.NewCronJob(
		description,
		slugURL,
		confFiles.TeresaYaml.Cron.Schedule,
		imgs,
		a,
		ops.fileStorage,
		strings.Split(confFiles.Procfile[a.ProcessType], " ")...,
	)

	if err := ops.k8s.CreateOrUpdateCronJob(cronSpec); err != nil {
		log.WithError(err).Errorf("Creating CronJob %s", a.Name)
		return err
	}
	fmt.Fprintln(w, fmt.Sprintf("The CronJob %s has been successfully deployed", a.Name))
	return nil
}

func (ops *DeployOperations) exposeApp(a *app.App, w io.Writer) error {
	if a.ProcessType != app.ProcessTypeWeb {
		return nil
	}
	svcType := ops.serviceType(a)
	if err := ops.k8s.ExposeDeploy(a.Name, a.Name, a.VirtualHost, svcType, a.Protocol, w); err != nil {
		return err
	}
	return nil // already exposed
}

func (ops *DeployOperations) podRun(ctx context.Context, podSpec *spec.Pod, stream io.Writer) error {
	podStream, runErrChan := ops.execOps.RunCommandBySpec(ctx, podSpec)
	go io.Copy(stream, podStream)

	if err := <-runErrChan; err != nil {
		if err == exec.ErrTimeout {
			return err
		}
		return ErrPodRunFail
	}

	return nil
}

func (ops *DeployOperations) List(user *database.User, appName string) ([]*ReplicaSetListItem, error) {
	if _, err := ops.appOps.Get(appName); err != nil {
		return nil, err
	}

	if !ops.appOps.HasPermission(user, appName) {
		return nil, auth.ErrPermissionDenied
	}

	items, err := ops.k8s.ReplicaSetListByLabel(appName, runLabel, appName)
	if err != nil {
		return nil, teresa_errors.NewInternalServerError(err)
	}

	return items, nil
}

func (ops *DeployOperations) Rollback(user *database.User, appName, revision string) error {
	a, err := ops.appOps.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}
	if err = ops.k8s.DeployRollbackToRevision(appName, appName, revision); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}
	env, err := ops.k8s.ContainerExplicitEnvVars(appName, appName, appName)
	if err != nil {
		return teresa_errors.NewInternalServerError(err)
	}
	appEnv := []*app.EnvVar{}
	for _, ev := range env {
		if isProtectedEnvVar(ev.Key) {
			continue
		}
		appEnv = append(appEnv, ev)
	}
	a.EnvVars = appEnv
	if err := ops.appOps.SaveApp(a, user.Email); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}
	return nil
}

func (ops *DeployOperations) serviceType(a *app.App) string {
	if a.Internal {
		return internalSvcType
	}
	return ops.opts.DefaultServiceType
}

func isProtectedEnvVar(key string) bool {
	for _, name := range slug.ProtectedEnvVars {
		if name == key {
			return true
		}
	}
	return false
}

func (ops *DeployOperations) watchDeploy(appName, deployId string, w io.Writer, errChan chan<- error) {
	fmt.Fprintln(w, "\nMonitoring rolling update...(hit Ctrl-C to quit)")
	if err := ops.k8s.WatchDeploy(appName, appName); err != nil {
		errChan <- err
		return
	}
	fmt.Fprintln(w, "Rolling update finished successfully")
}

func NewDeployOperations(aOps app.Operations, k8s K8sOperations, s storage.Storage, execOps exec.Operations, buildOps build.Operations, opts *Options) Operations {
	return &DeployOperations{
		appOps:      aOps,
		k8s:         k8s,
		fileStorage: s,
		execOps:     execOps,
		opts:        opts,
		buildOps:    buildOps,
	}
}
