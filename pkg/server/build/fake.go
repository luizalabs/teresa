package build

import (
	"io"
	"time"

	context "golang.org/x/net/context"

	"github.com/luizalabs/teresa/pkg/server/database"
)

type FakeOperations struct {
}

func (f *FakeOperations) Create(ctx context.Context, appName string, buildName string, u *database.User, tarBall io.ReadSeeker, runApp bool) (io.ReadCloser, <-chan error) {
	return nil, make(chan error)
}

func (f *FakeOperations) CreateByOpts(ctx context.Context, opt *CreateOptions) error {
	return nil
}

func (f *FakeOperations) List(appName string, u *database.User) ([]*Build, error) {
	return []*Build{
		&Build{Name: "v1.0.0", LastModified: time.Now()},
		&Build{Name: "v1.1.0-rc1", LastModified: time.Now()},
	}, nil
}

func (f *FakeOperations) Run(ctx context.Context, appName, buildName string, u *database.User) (io.ReadCloser, <-chan error) {
	return nil, make(chan error)
}

func (f *FakeOperations) Delete(appName, buildName string, u *database.User) error {
	return nil
}

func NewFakeOperations() *FakeOperations {
	return new(FakeOperations)
}
