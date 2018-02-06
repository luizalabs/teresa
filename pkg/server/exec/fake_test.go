package exec

import (
	"bufio"
	"fmt"
	"io"
	"testing"

	context "golang.org/x/net/context"
)

type commandCall func() (io.ReadCloser, <-chan error)

func TestFakeOperationsCommandFuncs(t *testing.T) {
	fake := &FakeOperations{}

	var testCases = []struct {
		cmdCall     commandCall
		expectedErr error
	}{
		{func() (io.ReadCloser, <-chan error) { return fake.RunCommand(context.Background(), nil, "") }, nil},
		{func() (io.ReadCloser, <-chan error) { return fake.RunCommand(context.Background(), nil, "") }, fmt.Errorf("some error")},
		{func() (io.ReadCloser, <-chan error) { return fake.RunCommandBySpec(context.Background(), nil) }, nil},
		{func() (io.ReadCloser, <-chan error) { return fake.RunCommandBySpec(context.Background(), nil) }, fmt.Errorf("some error")},
	}

	for _, tc := range testCases {
		fake.ExpectedErr = tc.expectedErr
		rc, errChan := tc.cmdCall()
		defer rc.Close()

		if err := <-errChan; err != tc.expectedErr {
			t.Errorf("expected %v, got %v", tc.expectedErr, err)
		}

		scanner := bufio.NewScanner(rc)
		if !scanner.Scan() {
			t.Fatal("cannot scan command output")
		}
		if err := scanner.Err(); err != nil {
			t.Fatal("error on read command output", err)
		}
		if txt := scanner.Text(); txt != "command output" {
			t.Errorf("expected command output, got %s", txt)
		}

	}
}
