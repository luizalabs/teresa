package deploy

import (
	"fmt"
	"io"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/app"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
	"github.com/pborman/uuid"
)

type Operations interface {
	Deploy(user *storage.User, appName string, tarBall io.ReadSeeker, description string) (io.ReadCloser, error)
}

type K8sOperations interface {
	PodRun(podSpec *PodSpec) (io.ReadCloser, <-chan int, error)
	CreateDeploy(deploySpec *DeploySpec) error
	NamespaceAnnotation(namespace, annotation string) (string, error)
}

type DeployOperations struct {
	appOps      app.Operations
	fileStorage st.Storage
	k8s         K8sOperations
}

func (ops *DeployOperations) Deploy(user *storage.User, appName string, tarBall io.ReadSeeker, description string) (io.ReadCloser, error) {
	a, err := ops.appOps.Meta(appName)
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

	tYaml, err := getTeresaYamlFromTarBall(tarBall) // get tYaml and parse before update deploy
	if err != nil {
		return nil, err
	}

	deployId := genDeployId()
	buildDest := fmt.Sprintf("deploys/%s/%s/out", appName, deployId)

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		if err = ops.buildApp(tarBall, a, deployId, buildDest, w); err != nil {
			return
		}
		ops.createDeploy(a, tYaml, description, buildDest)
	}()
	return r, nil
}

func (ops *DeployOperations) createDeploy(a *app.App, tYaml *TeresaYaml, description, slugPath string) error {
	deploySpec := newDeploySpec(a, tYaml, ops.fileStorage, description, slugPath, a.ProcessType)
	return ops.k8s.CreateDeploy(deploySpec)
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

	io.Copy(stream, podStream)
	exitCode := <-exitCodeChan
	if exitCode != 0 {
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
