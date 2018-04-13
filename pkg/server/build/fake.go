package build

import (
	"io"

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

func NewFakeOperations() *FakeOperations {
	return new(FakeOperations)
}
