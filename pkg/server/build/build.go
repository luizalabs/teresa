package build

import (
	"fmt"
	"io"

	context "golang.org/x/net/context"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/exec"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

type Operations interface {
	CreateByOpts(ctx context.Context, opts *CreateOptions) error
}

type Options struct {
	SlugBuilderImage string
	BuildLimitCPU    string
	BuildLimitMemory string
}

type BuildOperations struct {
	fileStorage storage.Storage
	execOps     exec.Operations
	opts        *Options
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

func (ops BuildOperations) CreateByOpts(ctx context.Context, opts *CreateOptions) error {
	opts.TarBall.Seek(0, 0)
	if err := ops.fileStorage.UploadFile(opts.SlugIn, opts.TarBall); err != nil {
		fmt.Fprintln(opts.Stream, "The Build failed to upload the TarBall to slug storage")
		return err
	}

	podSpec := spec.NewBuilder(
		fmt.Sprintf("build-%s", opts.BuildName),
		opts.SlugIn,
		opts.SlugDest,
		ops.opts.SlugBuilderImage,
		opts.App,
		ops.fileStorage,
		&spec.ContainerLimits{
			CPU:    ops.opts.BuildLimitCPU,
			Memory: ops.opts.BuildLimitMemory,
		},
	)

	podStream, runErrChan := ops.execOps.RunCommandBySpec(ctx, podSpec)
	go io.Copy(opts.Stream, podStream)

	if err := <-runErrChan; err != nil {
		if err == exec.ErrTimeout {
			return err
		}
		return ErrBuildFail
	}

	return nil
}

func NewBuildOperations(s storage.Storage, e exec.Operations, o *Options) *BuildOperations {
	return &BuildOperations{fileStorage: s, execOps: e, opts: o}
}
