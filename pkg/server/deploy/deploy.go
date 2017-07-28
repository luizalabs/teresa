package deploy

import (
	"fmt"
	"io"

	log "github.com/Sirupsen/logrus"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/app"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
	"github.com/luizalabs/teresa-api/pkg/server/teresa_errors"
	"github.com/pborman/uuid"
)

const ProcfileReleaseCmd = "release"

type Operations interface {
	Deploy(user *storage.User, appName string, tarBall io.ReadSeeker, description string, rhl int) (io.ReadCloser, error)
}

type K8sOperations interface {
	PodRun(podSpec *PodSpec) (io.ReadCloser, <-chan int, error)
	CreateOrUpdateDeploy(deploySpec *DeploySpec) error
	HasService(namespace, name string) (bool, error)
	CreateService(namespace, name string) error
}

type DeployOperations struct {
	appOps      app.Operations
	fileStorage st.Storage
	k8s         K8sOperations
}

func (ops *DeployOperations) Deploy(user *storage.User, appName string, tarBall io.ReadSeeker, description string, rhl int) (io.ReadCloser, error) {
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

	confFiles, err := getDeployConfigFilesFromTarBall(tarBall)
	if err != nil {
		return nil, teresa_errors.New(ErrInvalidTeresaYamlFile, err)
	}

	deployId := genDeployId()
	buildDest := fmt.Sprintf("deploys/%s/%s/out", appName, deployId)

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		if err = ops.buildApp(tarBall, a, deployId, buildDest, w); err != nil {
			log.WithError(err).Errorf("Building app %s", appName)
			return
		}

		slugURL := fmt.Sprintf("%s/slug.tgz", buildDest)
		releaseCmd := confFiles.Procfile[ProcfileReleaseCmd]
		if confFiles.Procfile != nil && releaseCmd != "" {
			if err := ops.runReleaseCmd(a, deployId, slugURL, w); err != nil {
				log.WithError(err).Errorf("Running release command %s in app %s", releaseCmd, appName)
				return
			}
		}

		if err := ops.createDeploy(a, confFiles.TeresaYaml, description, slugURL, rhl); err != nil {
			log.WithError(err).Errorf("Creating deploy app %s", appName)
			return
		}

		if err := ops.exposeService(a, w); err != nil {
			log.WithError(err).Errorf("Exposing service %s", appName)
		}
		fmt.Fprintln(w, fmt.Sprintf("The app %s has been successfully deployed", appName))
	}()
	return r, nil
}

func (ops *DeployOperations) runReleaseCmd(a *app.App, deployId, slugPath string, stream io.Writer) error {
	runCommandSpec := newRunCommandSpec(a, deployId, ProcfileReleaseCmd, slugPath, ops.fileStorage)

	fmt.Fprintln(stream, "Running release command")
	podStream, exitCodeChan, err := ops.k8s.PodRun(runCommandSpec)
	if err != nil {
		return err
	}
	go io.Copy(stream, podStream)

	exitCode, ok := <-exitCodeChan
	if !ok || exitCode != 0 {
		return ErrReleaseFail
	}
	return nil
}

func (ops *DeployOperations) createDeploy(a *app.App, tYaml *TeresaYaml, description, slugPath string, rhl int) error {
	deploySpec := newDeploySpec(a, tYaml, ops.fileStorage, description, slugPath, a.ProcessType, rhl)
	return ops.k8s.CreateOrUpdateDeploy(deploySpec)
}

func (ops *DeployOperations) exposeService(a *app.App, w io.Writer) error {
	if a.ProcessType != app.ProcessTypeWeb {
		return nil
	}
	hasSrv, err := ops.k8s.HasService(a.Name, a.Name)
	if err != nil {
		return err
	}
	if !hasSrv {
		fmt.Fprintln(w, "Exposing service")
		if err := ops.k8s.CreateService(a.Name, a.Name); err != nil {
			return err
		}
	}
	return nil // already exposed
}

func (ops *DeployOperations) buildApp(tarBall io.ReadSeeker, a *app.App, deployId, buildDest string, stream io.Writer) error {
	tarBall.Seek(0, 0)
	tarBallLocation := fmt.Sprintf("deploys/%s/%s/in/app.tar.gz", a.Name, deployId)
	if err := ops.fileStorage.UploadFile(tarBallLocation, tarBall); err != nil {
		return err
	}
	buildSpec := newBuildSpec(a, deployId, tarBallLocation, buildDest, ops.fileStorage)
	podStream, exitCodeChan, err := ops.k8s.PodRun(buildSpec)
	if err != nil {
		return err
	}
	go io.Copy(stream, podStream)

	exitCode, ok := <-exitCodeChan
	if !ok || exitCode != 0 {
		return ErrBuildFail
	}
	return nil
}

func genDeployId() string {
	return uuid.New()[:8]
}

func NewDeployOperations(aOps app.Operations, k8s K8sOperations, s st.Storage) Operations {
	return &DeployOperations{appOps: aOps, k8s: k8s, fileStorage: s}
}
