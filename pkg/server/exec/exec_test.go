package exec

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

type fakeK8sOperations struct {
	errDeployAnnotation error
	errPodRun           error
	isNotFound          bool
	exitCodePodRun      int
}

func (f *fakeK8sOperations) DeployAnnotation(namespace string, deployName string, annotation string) (string, error) {
	return "slug", f.errDeployAnnotation
}

func (f *fakeK8sOperations) PodRun(podSpec *spec.Pod) (io.ReadCloser, <-chan int, error) {
	r := bytes.NewBufferString("foo\nbar")

	exitCodeChan := make(chan int, 1)
	exitCodeChan <- f.exitCodePodRun

	return ioutil.NopCloser(r), exitCodeChan, f.errPodRun
}

func (f *fakeK8sOperations) IsNotFound(err error) bool {
	return f.isNotFound
}

func TestOpsCommand(t *testing.T) {
	k8sOps := &fakeK8sOperations{}
	ops := NewOperations(app.NewFakeOperations(), k8sOps, storage.NewFake(), &Defaults{})

	rc, errChan := ops.Command(&database.User{}, "teresa", "ls")
	defer rc.Close()

	if err := <-errChan; err != nil {
		t.Errorf("expected non error, got %v", err)
	}
}

func TestOpsCommandAppNotFound(t *testing.T) {
	ops := NewOperations(app.NewFakeOperations(), &fakeK8sOperations{}, storage.NewFake(), &Defaults{})
	_, errChan := ops.Command(&database.User{}, "notfound", "ls")
	if err := <-errChan; err != app.ErrNotFound {
		t.Errorf("expected app.ErrNotFound, got %v", err)
	}
}

func TestOpsCommandPermissionDenied(t *testing.T) {
	ops := NewOperations(app.NewFakeOperations(), &fakeK8sOperations{}, storage.NewFake(), &Defaults{})
	_, errChan := ops.Command(&database.User{Email: "bad-user@luizalabs.com"}, "teresa", "ls")
	if err := <-errChan; err != auth.ErrPermissionDenied {
		t.Errorf("expected auth.ErrPermissionDenied, got %v", err)
	}
}

func TestOpsCommandDeployNotFound(t *testing.T) {
	k8sOps := &fakeK8sOperations{errDeployAnnotation: fmt.Errorf("not found"), isNotFound: true}
	ops := NewOperations(app.NewFakeOperations(), k8sOps, storage.NewFake(), &Defaults{})
	_, errChan := ops.Command(&database.User{}, "teresa", "ls")
	if err := <-errChan; err != ErrDeployNotFound {
		t.Errorf("expected ErrDeployNotFound, got %v", err)
	}
}

func TestOpsCommandBySpec(t *testing.T) {
	k8sOps := &fakeK8sOperations{}
	ops := NewOperations(app.NewFakeOperations(), k8sOps, storage.NewFake(), &Defaults{})

	rc, errChan := ops.CommandBySpec(&spec.Pod{})
	defer rc.Close()

	if err := <-errChan; err != nil {
		t.Errorf("expected non error, got %v", err)
	}
}

func TestCommandBySpecNoZeroExitCode(t *testing.T) {
	k8sOps := &fakeK8sOperations{exitCodePodRun: 1}
	ops := NewOperations(app.NewFakeOperations(), k8sOps, storage.NewFake(), &Defaults{})

	rc, errChan := ops.CommandBySpec(&spec.Pod{})
	defer rc.Close()

	if err := <-errChan; err != ErrNonZeroExitCode {
		t.Errorf("expected ErrNonZeroExitCode, got %v", err)
	}
}
