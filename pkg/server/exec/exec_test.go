package exec

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
	"time"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"github.com/luizalabs/teresa/pkg/server/storage"
	context "golang.org/x/net/context"
)

type fakeK8sOperations struct {
	errDeployAnnotation error
	errPodRun           error
	isNotFound          bool
	exitCodePodRun      int
	podRunDelay         int
}

func (f *fakeK8sOperations) DeployAnnotation(namespace string, deployName string, annotation string) (string, error) {
	return "slug", f.errDeployAnnotation
}

func (f *fakeK8sOperations) PodRun(podSpec *spec.Pod) (io.ReadCloser, <-chan int, error) {
	r := bytes.NewBufferString("foo\nbar")

	exitCodeChan := make(chan int)
	go func() {
		time.Sleep(time.Duration(f.podRunDelay) * time.Millisecond)
		exitCodeChan <- f.exitCodePodRun
	}()

	return ioutil.NopCloser(r), exitCodeChan, f.errPodRun
}

func (f *fakeK8sOperations) IsNotFound(err error) bool {
	return f.isNotFound
}

func (f *fakeK8sOperations) DeletePod(namespace, podName string) error {
	return nil
}

func TestOpsRunCommand(t *testing.T) {
	k8sOps := &fakeK8sOperations{}
	ops := NewOperations(app.NewFakeOperations(), k8sOps, storage.NewFake(), &Defaults{})

	rc, errChan := ops.RunCommand(context.Background(), &database.User{}, "teresa", "ls")
	defer rc.Close()

	if err := <-errChan; err != nil {
		t.Errorf("expected non error, got %v", err)
	}
}

func TestOpsRunCommandAppNotFound(t *testing.T) {
	ops := NewOperations(app.NewFakeOperations(), &fakeK8sOperations{}, storage.NewFake(), &Defaults{})
	_, errChan := ops.RunCommand(context.Background(), &database.User{}, "notfound", "ls")
	if err := <-errChan; err != app.ErrNotFound {
		t.Errorf("expected app.ErrNotFound, got %v", err)
	}
}

func TestOpsRunCommandPermissionDenied(t *testing.T) {
	ops := NewOperations(app.NewFakeOperations(), &fakeK8sOperations{}, storage.NewFake(), &Defaults{})
	_, errChan := ops.RunCommand(context.Background(), &database.User{Email: "bad-user@luizalabs.com"}, "teresa", "ls")
	if err := <-errChan; err != auth.ErrPermissionDenied {
		t.Errorf("expected auth.ErrPermissionDenied, got %v", err)
	}
}

func TestOpsRunCommandDeployNotFound(t *testing.T) {
	k8sOps := &fakeK8sOperations{errDeployAnnotation: fmt.Errorf("not found"), isNotFound: true}
	ops := NewOperations(app.NewFakeOperations(), k8sOps, storage.NewFake(), &Defaults{})
	_, errChan := ops.RunCommand(context.Background(), &database.User{}, "teresa", "ls")
	if err := <-errChan; err != ErrDeployNotFound {
		t.Errorf("expected ErrDeployNotFound, got %v", err)
	}
}

func TestOpsRunCommandBySpec(t *testing.T) {
	k8sOps := &fakeK8sOperations{}
	ops := NewOperations(app.NewFakeOperations(), k8sOps, storage.NewFake(), &Defaults{})

	rc, errChan := ops.RunCommandBySpec(context.Background(), &spec.Pod{})
	defer rc.Close()

	if err := <-errChan; err != nil {
		t.Errorf("expected non error, got %v", err)
	}
}

func TestRunCommandBySpecNoZeroExitCode(t *testing.T) {
	k8sOps := &fakeK8sOperations{exitCodePodRun: ExitCodeError}
	ops := NewOperations(app.NewFakeOperations(), k8sOps, storage.NewFake(), &Defaults{})

	rc, errChan := ops.RunCommandBySpec(context.Background(), &spec.Pod{})
	defer rc.Close()

	if err := <-errChan; err != ErrNonZeroExitCode {
		t.Errorf("expected ErrNonZeroExitCode, got %v", err)
	}
}

func TestRunCommandBySpecTimeoutExitCode(t *testing.T) {
	k8sOps := &fakeK8sOperations{exitCodePodRun: ExitCodeTimeout}
	ops := NewOperations(app.NewFakeOperations(), k8sOps, storage.NewFake(), &Defaults{})

	rc, errChan := ops.RunCommandBySpec(context.Background(), &spec.Pod{})
	defer rc.Close()

	if err := <-errChan; err != ErrTimeout {
		t.Errorf("expected ErrTimeout, got %v", err)
	}
}

func TestOpsRunCommandBySpecContextCancelation(t *testing.T) {
	k8sOps := &fakeK8sOperations{podRunDelay: 10}
	ops := NewOperations(app.NewFakeOperations(), k8sOps, storage.NewFake(), &Defaults{})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rc, errChan := ops.RunCommandBySpec(ctx, &spec.Pod{})
	defer rc.Close()

	if err := <-errChan; err != context.Canceled {
		t.Errorf("expected context canceled, got %v", err)
	}
}
