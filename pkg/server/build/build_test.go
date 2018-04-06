package build

import (
	"bytes"
	"context"
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/exec"
	"github.com/luizalabs/teresa/pkg/server/storage"
	"github.com/luizalabs/teresa/pkg/server/test"
)

func TestCreateByOpts(t *testing.T) {
	var testCases = []struct {
		commandErr  error
		expectedErr error
	}{
		{nil, nil},
		{exec.ErrNonZeroExitCode, ErrBuildFail},
		{exec.ErrTimeout, exec.ErrTimeout},
	}

	for _, tc := range testCases {
		fakeExec := exec.NewFakeOperations()
		fakeExec.ExpectedErr = tc.commandErr

		ops := NewBuildOperations(storage.NewFake(), fakeExec, &Options{})
		err := ops.CreateByOpts(context.Background(), &CreateOptions{
			App:     &app.App{},
			TarBall: &test.FakeReadSeeker{},
			Stream:  new(bytes.Buffer),
		})

		if err != tc.expectedErr {
			t.Errorf("expected %v, got %v", tc.expectedErr, err)
		}
	}
}
