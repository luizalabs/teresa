package deploy

import (
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	st "github.com/luizalabs/teresa/pkg/server/storage"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
	"github.com/pborman/uuid"
)

const (
	ProcfileReleaseCmd = "release"
	runLabel           = "run"
)

type Operations interface {
	Deploy(user *database.User, appName string, tarBall io.ReadSeeker, description string, opts *Options) (io.ReadCloser, error)
	List(user *database.User, appName string) ([]*ReplicaSetListItem, error)
	Rollback(user *database.User, appName, revision string) error
}

type K8sOperations interface {
	PodRun(podSpec *PodSpec) (io.ReadCloser, <-chan int, error)
	CreateOrUpdateDeploy(deploySpec *DeploySpec) error
	ExposeDeploy(namespace, name, vHost string, w io.Writer) error
	ReplicaSetListByLabel(namespace, label, value string) ([]*ReplicaSetListItem, error)
	DeployRollbackToRevision(namespace, name, revision string) error
}

type DeployOperations struct {
	appOps      app.Operations
	fileStorage st.Storage
	k8s         K8sOperations
}

func (ops *DeployOperations) Deploy(user *database.User, appName string, tarBall io.ReadSeeker, description string, opts *Options) (io.ReadCloser, error) {
	a, err := ops.appOps.Get(appName)
	if err != nil {
		return nil, err
	}

	teamName, err := ops.appOps.TeamName(appName)
	if err != nil {
		return nil, err
	}
	a.Team = teamName

	if !ops.appOps.HasPermission(user, appName) {
		return nil, auth.ErrPermissionDenied
	}

	confFiles, err := getDeployConfigFilesFromTarBall(tarBall, a.ProcessType)
	if err != nil {
		return nil, teresa_errors.New(ErrInvalidTeresaYamlFile, err)
	}

	deployId := genDeployId()
	buildDest := fmt.Sprintf("deploys/%s/%s/out", appName, deployId)

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		if err = ops.buildApp(tarBall, a, deployId, buildDest, w, opts); err != nil {
			log.WithError(err).WithField("id", deployId).Errorf("Building app %s", appName)
			return
		}

		slugURL := fmt.Sprintf("%s/slug.tgz", buildDest)
		releaseCmd := confFiles.Procfile[ProcfileReleaseCmd]
		if confFiles.Procfile != nil && releaseCmd != "" {
			if err := ops.runReleaseCmd(a, deployId, slugURL, w, opts); err != nil {
				log.WithError(err).WithField("id", deployId).Errorf("Running release command %s in app %s", releaseCmd, appName)
				return
			}
		}

		if err := ops.createDeploy(a, confFiles.TeresaYaml, description, slugURL, opts); err != nil {
			log.WithError(err).Errorf("Creating deploy app %s", appName)
			return
		}

		if err := ops.exposeApp(a, w); err != nil {
			log.WithError(err).Errorf("Exposing service %s", appName)
		}
		fmt.Fprintln(w, fmt.Sprintf("The app %s has been successfully deployed", appName))
	}()
	return r, nil
}

func (ops *DeployOperations) runReleaseCmd(a *app.App, deployId, slugPath string, stream io.Writer, opts *Options) error {
	runCommandSpec := newRunCommandSpec(a, deployId, ProcfileReleaseCmd, slugPath, ops.fileStorage, opts)

	fmt.Fprintln(stream, "Running release command")
	err := ops.podRun(runCommandSpec, stream)
	if err != nil {
		if err == ErrPodRunFail {
			return ErrReleaseFail
		}
		return err
	}
	return nil
}

func (ops *DeployOperations) createDeploy(a *app.App, tYaml *TeresaYaml, description, slugPath string, opts *Options) error {
	deploySpec := newDeploySpec(a, tYaml, ops.fileStorage, description, slugPath, a.ProcessType, opts)
	return ops.k8s.CreateOrUpdateDeploy(deploySpec)
}

func (ops *DeployOperations) exposeApp(a *app.App, w io.Writer) error {
	if a.ProcessType != app.ProcessTypeWeb {
		return nil
	}
	if err := ops.k8s.ExposeDeploy(a.Name, a.Name, a.VirtualHost, w); err != nil {
		return err
	}
	return nil // already exposed
}

func (ops *DeployOperations) buildApp(tarBall io.ReadSeeker, a *app.App, deployId, buildDest string, stream io.Writer, opts *Options) error {
	tarBall.Seek(0, 0)
	tarBallLocation := fmt.Sprintf("deploys/%s/%s/in/app.tar.gz", a.Name, deployId)
	if err := ops.fileStorage.UploadFile(tarBallLocation, tarBall); err != nil {
		fmt.Fprintln(stream, "The Deploy failed to upload the tarBall to slug storage")
		return err
	}
	buildSpec := newBuildSpec(a, deployId, tarBallLocation, buildDest, ops.fileStorage, opts)
	err := ops.podRun(buildSpec, stream)
	if err != nil {
		if err == ErrPodRunFail {
			return ErrBuildFail
		}
		return err
	}
	return nil
}

func (ops *DeployOperations) podRun(podSpec *PodSpec, stream io.Writer) error {
	podStream, exitCodeChan, err := ops.k8s.PodRun(podSpec)
	if err != nil {
		return err
	}
	go io.Copy(stream, podStream)

	exitCode, ok := <-exitCodeChan
	if !ok || exitCode != 0 {
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
	app, err := ops.appOps.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}

	if err = ops.k8s.DeployRollbackToRevision(appName, appName, revision); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	if err := ops.appOps.SaveApp(app, user.Email); err != nil {
		return teresa_errors.NewInternalServerError(err)
	}

	return nil
}

func genDeployId() string {
	return uuid.New()[:8]
}

func NewDeployOperations(aOps app.Operations, k8s K8sOperations, s st.Storage) Operations {
	return &DeployOperations{appOps: aOps, k8s: k8s, fileStorage: s}
}
