package exec

import (
	"fmt"
	"io"

	context "golang.org/x/net/context"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"github.com/luizalabs/teresa/pkg/server/storage"
	"github.com/luizalabs/teresa/pkg/server/uid"
)

type Operations interface {
	RunCommand(ctx context.Context, user *database.User, appName string, command ...string) (io.ReadCloser, <-chan error)
	RunCommandBySpec(ctx context.Context, podSpec *spec.Pod) (io.ReadCloser, <-chan error)
}

type K8sOperations interface {
	DeployAnnotation(namespace, deployName, annotation string) (string, error)
	PodRun(podSpec *spec.Pod) (io.ReadCloser, <-chan int, error)
	IsNotFound(err error) bool
	DeletePod(namespace, podName string) error
}

type Defaults struct {
	RunnerImage  string
	LimitsCPU    string
	LimitsMemory string
}

type ExecOperations struct {
	appOps   app.Operations
	fs       storage.Storage
	k8s      K8sOperations
	defaults *Defaults
}

func (ops *ExecOperations) RunCommand(ctx context.Context, user *database.User, appName string, command ...string) (io.ReadCloser, <-chan error) {
	errChan := make(chan error, 1)
	a, err := ops.appOps.CheckPermAndGet(user, appName)
	if err != nil {
		errChan <- err
		return nil, errChan
	}

	currentSlug, err := ops.k8s.DeployAnnotation(a.Name, a.Name, spec.SlugAnnotation)
	if err != nil {
		if ops.k8s.IsNotFound(err) {
			errChan <- ErrDeployNotFound
		} else {
			errChan <- err
		}
		return nil, errChan
	}

	podSpec := spec.NewRunner(
		fmt.Sprintf("exec-command-%s-%s", appName, uid.New()),
		currentSlug,
		ops.defaults.RunnerImage,
		a,
		ops.fs,
		&spec.ContainerLimits{
			CPU:    ops.defaults.LimitsCPU,
			Memory: ops.defaults.LimitsMemory,
		},
		command...,
	)

	return ops.RunCommandBySpec(ctx, podSpec)
}

func (ops *ExecOperations) RunCommandBySpec(ctx context.Context, podSpec *spec.Pod) (io.ReadCloser, <-chan error) {
	errChan := make(chan error)
	r, w := io.Pipe()
	go func() {
		defer func() {
			w.Close()
			close(errChan)
		}()

		podStream, exitCodeChain, err := ops.k8s.PodRun(podSpec)
		if err != nil {
			errChan <- err
			return
		}
		go io.Copy(w, podStream)

		select {
		case <-ctx.Done():
			go ops.k8s.DeletePod(podSpec.Namespace, podSpec.Name)
			errChan <- ctx.Err()
		case ec := <-exitCodeChain:
			if ec != 0 {
				errChan <- ErrNonZeroExitCode
			}
		}
	}()
	return r, errChan
}

func NewOperations(appOps app.Operations, k8s K8sOperations, fs storage.Storage, defaults *Defaults) Operations {
	return &ExecOperations{
		appOps:   appOps,
		fs:       fs,
		k8s:      k8s,
		defaults: defaults,
	}
}
