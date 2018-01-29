package exec

import (
	"fmt"
	"io"

	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/spec"
)

type FakeOperations struct {
	ExpectedErr error
}

func (f *FakeOperations) Command(user *database.User, appName string, command ...string) (io.ReadCloser, <-chan error) {
	return f.CommandBySpec(nil)
}

func (f *FakeOperations) CommandBySpec(podSpec *spec.Pod) (io.ReadCloser, <-chan error) {
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
