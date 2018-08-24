package deploy

import (
	"fmt"
	"io"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	context "golang.org/x/net/context"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/build"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/exec"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"github.com/luizalabs/teresa/pkg/server/storage"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
	"github.com/luizalabs/teresa/pkg/server/uid"
	"github.com/luizalabs/teresa/pkg/server/validation"
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
	ExposeDeploy(namespace, name, svcType, portName string, vHosts []string, w io.Writer) error
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
	a, err := ops.appOps.CheckPermAndGet(user, appName)
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

	confFiles, err := getDeployConfigFilesFromTarBall(tarBall, a.Name, a.ProcessType)
	if err != nil {
		errChan <- teresa_errors.New(ErrInvalidTeresaYamlFile, err)
		return nil, errChan
	}

	deployId := uid.New()
	buildIn := fmt.Sprintf("deploys/%s/%s/in/app.tgz", a.Name, deployId)
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

		if err := ops.appOps.SaveApp(a, user.Email); err != nil {
			log.WithError(err).WithField("id", deployId).Errorf("Saving last deploy user (%s) of app %s", user.Name, appName)
		}

		if !app.IsCronJob(a.ProcessType) {
			ops.watchDeploy(appName, deployId, w, errChan)
		}
	}()
	return r, errChan
}

func (ops *DeployOperations) runReleaseCmd(a *app.App, deployId, slugURL string, stream io.Writer) error {
	podName := fmt.Sprintf("release-%s-%s", a.Name, deployId)
	podSpec := spec.NewRunnerPodBuilder(podName, ops.opts.SlugRunnerImage, ops.opts.SlugStoreImage).
		ForApp(a).
		WithSlug(slugURL).
		WithLimits(ops.opts.BuildLimitCPU, ops.opts.BuildLimitMemory).
		WithStorage(ops.fileStorage).
		WithArgs([]string{"start", ProcfileReleaseCmd}).
		Build()

	fmt.Fprintln(stream, "Running release command")
	if err := ops.podRun(context.Background(), podSpec, stream); err != nil {
		if err == ErrPodRunFail {
			return ErrReleaseFail
		}
		return err
	}
	return nil
}

func (ops *DeployOperations) createOrUpdateDeploy(a *app.App, confFiles *DeployConfigFiles, w io.Writer, slugURL, description, deployId string) error {
	if releaseCmd := confFiles.Procfile[ProcfileReleaseCmd]; releaseCmd != "" {
		if err := ops.runReleaseCmd(a, deployId, slugURL, w); err != nil {
			log.WithError(err).WithField("id", deployId).Errorf("Running release command %s in app %s", releaseCmd, a.Name)
			return err
		}
	}
	labels := map[string]string{"run": a.Name}
	podBuilder := spec.NewRunnerPodBuilder(a.Name, ops.opts.SlugRunnerImage, ops.opts.SlugStoreImage).
		ForApp(a).
		WithSlug(slugURL).
		WithLabels(labels).
		WithStorage(ops.fileStorage).
		WithArgs([]string{"start", a.ProcessType})

	if confFiles.NginxConf != "" && app.IsWebApp(a.ProcessType) {
		data := map[string]string{spec.NginxConfFile: confFiles.NginxConf}
		if err := ops.k8s.CreateOrUpdateConfigMap(a.Name, a.Name, data); err != nil {
			log.WithError(err).Errorf("Creating config to nginx of app %s", a.Name)
			return err
		}
		podBuilder = podBuilder.WithNginxSideCar(ops.opts.NginxImage)
	} else {
		err := ops.k8s.DeleteConfigMap(a.Name, a.Name)
		if err != nil && !ops.k8s.IsNotFound(err) {
			return err
		}
	}

	csp, err := spec.NewCloudSQLProxy(ops.opts.CloudSQLProxyImage, confFiles.TeresaYaml)
	if err != nil {
		return errors.Wrap(err, "failed to create the deploy")
	}
	podBuilder = podBuilder.WithCloudSQLProxySideCar(csp)

	deploySpec := spec.NewDeployBuilder(slugURL).
		WithPod(podBuilder.Build()).
		WithDescription(description).
		WithRevisionHistoryLimit(ops.opts.RevisionHistoryLimit).
		WithTeresaYaml(confFiles.TeresaYaml).
		WithMatchLabels(labels).
		Build()

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
	if confFiles.TeresaYaml == nil || confFiles.TeresaYaml.Cron == nil {
		return ErrCronScheduleNotFound
	}

	podSpec := spec.NewRunnerPodBuilder(a.Name, ops.opts.SlugRunnerImage, ops.opts.SlugStoreImage).
		ForApp(a).
		WithSlug(slugURL).
		WithStorage(ops.fileStorage).
		WithArgs(strings.Split(confFiles.Procfile[a.ProcessType], " ")).
		Build()

	cronSpec := spec.NewCronJobBuilder(slugURL).
		WithPod(podSpec).
		WithDescription(description).
		WithSchedule(confFiles.TeresaYaml.Cron.Schedule).
		Build()

	if err := ops.k8s.CreateOrUpdateCronJob(cronSpec); err != nil {
		log.WithError(err).Errorf("Creating CronJob %s", a.Name)
		return err
	}
	fmt.Fprintln(w, fmt.Sprintf("The CronJob %s has been successfully deployed", a.Name))
	return nil
}

func (ops *DeployOperations) exposeApp(a *app.App, w io.Writer) error {
	if !app.IsWebApp(a.ProcessType) {
		return nil
	}
	svcType := ops.serviceType(a)
	vHosts := strings.Split(a.VirtualHost, ",")
	if err := ops.k8s.ExposeDeploy(a.Name, a.Name, svcType, a.Protocol, vHosts, w); err != nil {
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
		if validation.IsProtectedEnvVar(ev.Key) {
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
