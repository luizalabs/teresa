package exec

import (
	"fmt"
	"io"

	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/spec"
	context "golang.org/x/net/context"
)

type FakeOperations struct {
	ExpectedErr error
}

func (f *FakeOperations) RunCommand(ctx context.Context, user *database.User, appName string, command ...string) (io.ReadCloser, <-chan error) {
	return f.RunCommandBySpec(ctx, nil)
}

func (f *FakeOperations) RunCommandBySpec(ctx context.Context, podSpec *spec.Pod) (io.ReadCloser, <-chan error) {
	errChan := make(chan error, 1)
	r, w := io.Pipe()
	go func() {
		defer w.Close()

		errChan <- f.ExpectedErr
		fmt.Fprintf(w, "command output")
	}()

	return r, errChan
}

func NewFakeOperations() *FakeOperations {
	return new(FakeOperations)
}
