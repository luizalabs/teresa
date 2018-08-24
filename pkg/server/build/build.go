package build

import (
	"fmt"
	"io"

	context "golang.org/x/net/context"

	log "github.com/Sirupsen/logrus"
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/exec"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

type Operations interface {
	CreateByOpts(ctx context.Context, opts *CreateOptions) error
	Create(ctx context.Context, appName, buildName string, u *database.User, tarBall io.ReadSeeker, runApp bool) (io.ReadCloser, <-chan error)
	List(appName string, u *database.User) ([]*Build, error)
	Run(ctx context.Context, appName, buildName string, u *database.User) (io.ReadCloser, <-chan error)
	Delete(appName, buildName string, u *database.User) error
}

type K8sOperations interface {
	CreateService(svcSpec *spec.Service) error
	DeletePod(namespace, name string) error
	DeleteService(namespace, name string) error
	WatchServiceURL(namespace, name string) ([]string, error)
	IsInvalid(err error) bool
}

type Options struct {
	SlugBuilderImage string
	SlugRunnerImage  string
	SlugStoreImage   string
	BuildLimitCPU    string
	BuildLimitMemory string
}

type BuildOperations struct {
	fileStorage storage.Storage
	execOps     exec.Operations
	appOps      app.Operations
	opts        *Options
	k8s         K8sOperations
	buildLimits *spec.ContainerLimits
}

// CreateOptions define arguments of method `CreateByOpts`
type CreateOptions struct {
	App       *app.App
	BuildName string
	SlugIn    string
	SlugDest  string
	TarBall   io.ReadSeeker
	Stream    io.Writer
}

func formatPodName(appName, buildName string) string {
	return fmt.Sprintf("run-%s-build-%s", appName, buildName)
}

func (ops *BuildOperations) Create(ctx context.Context, appName, buildName string, u *database.User, tarBall io.ReadSeeker, runApp bool) (io.ReadCloser, <-chan error) {
	errChan := make(chan error, 1)

	a, err := ops.appOps.CheckPermAndGet(u, appName)
	if err != nil {
		errChan <- err
		return nil, errChan
	}

	r, w := io.Pipe()
	go func() {
		defer w.Close()

		err = ops.CreateByOpts(ctx, &CreateOptions{
			App:       a,
			BuildName: buildName,
			SlugIn:    fmt.Sprintf("builds/%s/%s/in/app.tgz", a.Name, buildName),
			SlugDest:  fmt.Sprintf("builds/%s/%s/out", a.Name, buildName),
			TarBall:   tarBall,
			Stream:    w,
		})
		if err != nil {
			errChan <- err
			return
		}

		if runApp {
			if err := ops.runInternal(ctx, a, buildName, w); err != nil {
				errChan <- err
			}
		}
	}()
	return r, errChan
}

func (ops *BuildOperations) CreateByOpts(ctx context.Context, opts *CreateOptions) error {
	opts.TarBall.Seek(0, 0)
	if err := ops.fileStorage.UploadFile(opts.SlugIn, opts.TarBall); err != nil {
		fmt.Fprintln(opts.Stream, "The Build failed to upload the TarBall to slug storage")
		return err
	}

	podName := fmt.Sprintf("build-%s", opts.BuildName)
	podSpec := spec.NewBuildPodBuilder(podName, ops.opts.SlugBuilderImage).
		ForApp(opts.App).
		WithTarBallPath(opts.SlugIn).
		SendSlugTo(opts.SlugDest).
		WithStorage(ops.fileStorage).
		WithLimits(ops.buildLimits.CPU, ops.buildLimits.Memory).
		Build()

	podStream, runErrChan := ops.execOps.RunCommandBySpec(ctx, podSpec)
	go io.Copy(opts.Stream, podStream)

	if err := <-runErrChan; err != nil {
		log.WithError(err).Errorf("failed to build app %s", opts.App.Name)
		if err == exec.ErrTimeout {
			return err
		} else if ops.k8s.IsInvalid(err) {
			return ErrInvalidBuildName
		}
		return ErrBuildFail
	}

	return nil
}

func (ops *BuildOperations) List(appName string, u *database.User) ([]*Build, error) {
	if _, err := ops.appOps.CheckPermAndGet(u, appName); err != nil {
		return nil, err
	}

	path := fmt.Sprintf("builds/%s/", appName)
	items, err := ops.fileStorage.List(path)
	if err != nil {
		return nil, err
	}

	builds := make([]*Build, len(items))
	for i := range items {
		builds[i] = &Build{
			Name:         items[i].Name,
			LastModified: items[i].LastModified,
		}
	}
	return builds, nil
}

func (ops *BuildOperations) Run(ctx context.Context, appName, buildName string, u *database.User) (io.ReadCloser, <-chan error) {
	errChan := make(chan error, 1)

	a, err := ops.appOps.CheckPermAndGet(u, appName)
	if err != nil {
		errChan <- err
		return nil, errChan
	}

	path := fmt.Sprintf("builds/%s/%s/", appName, buildName)
	items, err := ops.fileStorage.List(path)
	if len(items) == 0 {
		errChan <- ErrInvalidBuildName
		return nil, errChan
	}

	r, w := io.Pipe()
	go func() {
		defer w.Close()
		if err := ops.runInternal(ctx, a, buildName, w); err != nil {
			errChan <- err
		}
	}()
	return r, errChan
}

func (ops *BuildOperations) Delete(appName, buildName string, u *database.User) error {
	if _, err := ops.appOps.CheckPermAndGet(u, appName); err != nil {
		return err
	}

	return ops.fileStorage.Delete(fmt.Sprintf("builds/%s/%s/", appName, buildName))
}

func (ops *BuildOperations) runInternal(ctx context.Context, a *app.App, buildName string, w io.Writer) error {
	slugURL := fmt.Sprintf("builds/%s/%s/out/slug.tgz", a.Name, buildName)
	podName := formatPodName(a.Name, buildName)
	podSpec := spec.NewRunnerPodBuilder(podName, ops.opts.SlugRunnerImage, ops.opts.SlugStoreImage).
		ForApp(a).
		WithSlug(slugURL).
		WithStorage(ops.fileStorage).
		WithLabels(map[string]string{"build": buildName}).
		WithArgs([]string{"start", a.ProcessType}).
		Build()

	if app.IsWebApp(a.ProcessType) {
		fmt.Fprintln(w, "\nExposing temporary service")
		url, err := ops.createService(a.Name, buildName, podSpec.Labels)
		if err != nil {
			return err
		}
		defer ops.k8s.DeleteService(a.Name, buildName)

		fmt.Fprintf(w, "Temporary URL: %s\n\n", url)
	}

	fmt.Fprintln(w, "Starting application")
	podStream, runErrChan := ops.execOps.RunCommandBySpec(ctx, podSpec)
	go io.Copy(w, podStream)
	defer ops.k8s.DeletePod(a.Name, formatPodName(a.Name, buildName))

	if err := <-runErrChan; err != nil {
		log.WithError(err).Errorf("failed to run app %s", a.Name)
		if err == exec.ErrTimeout {
			return err
		}
		return ErrBuildFail
	}
	return nil
}

func (ops *BuildOperations) createService(appName, buildName string, labels map[string]string) (string, error) {
	svcSpec := spec.NewService(
		appName,
		buildName,
		"LoadBalancer",
		[]spec.ServicePort{*spec.NewDefaultServicePort("")},
		labels,
	)
	if err := ops.k8s.CreateService(svcSpec); err != nil {
		return "", err
	}

	urls, err := ops.k8s.WatchServiceURL(appName, buildName)
	if err != nil {
		return "", err
	}
	return urls[0], err
}

func NewBuildOperations(s storage.Storage, a app.Operations, e exec.Operations, k K8sOperations, o *Options) *BuildOperations {
	bl := &spec.ContainerLimits{CPU: o.BuildLimitCPU, Memory: o.BuildLimitMemory}
	return &BuildOperations{
		appOps:      a,
		fileStorage: s,
		execOps:     e,
		k8s:         k,
		buildLimits: bl,
		opts:        o,
	}
}
