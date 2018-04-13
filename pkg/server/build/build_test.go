package build

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/exec"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"github.com/luizalabs/teresa/pkg/server/storage"
	"github.com/luizalabs/teresa/pkg/server/test"
)

type fakeK8sOperations struct {
	createServiceWasCalled bool
}

func (f *fakeK8sOperations) CreateService(svcSpec *spec.Service) error {
	f.createServiceWasCalled = true
	return nil
}

func (f *fakeK8sOperations) DeletePod(namespace string, name string) error {
	return nil
}

func (f *fakeK8sOperations) DeleteService(namespace string, name string) error {
	return nil
}

func (f *fakeK8sOperations) WatchServiceURL(namespace string, name string) ([]string, error) {
	return []string{"url1", "url2"}, nil
}

func consumeReader(rc io.ReadCloser) {
	for {
		b := make([]byte, 64)
		if _, err := rc.Read(b); err != nil {
			break
		}
	}
}

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

		ops := NewBuildOperations(
			storage.NewFake(),
			app.NewFakeOperations(),
			fakeExec,
			&fakeK8sOperations{},
			&Options{},
		)
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

func TestCreateAppNotFound(t *testing.T) {
	ops := NewBuildOperations(
		storage.NewFake(),
		app.NewFakeOperations(),
		exec.NewFakeOperations(),
		&fakeK8sOperations{},
		&Options{},
	)
	u := &database.User{}
	ctx := context.Background()
	_, errChan := ops.Create(ctx, "bad-app", "foo", u, &test.FakeReadSeeker{}, false)
	if err := <-errChan; err != app.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCreatePermissionDenied(t *testing.T) {
	ops := NewBuildOperations(
		storage.NewFake(),
		app.NewFakeOperations(),
		exec.NewFakeOperations(),
		&fakeK8sOperations{},
		&Options{},
	)
	u := &database.User{Email: "bad-user@luizalabs.com"}
	ctx := context.Background()
	_, errChan := ops.Create(ctx, "teresa", "test", u, &test.FakeReadSeeker{}, false)
	if err := <-errChan; err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestCreate(t *testing.T) {
	ops := NewBuildOperations(
		storage.NewFake(),
		app.NewFakeOperations(),
		exec.NewFakeOperations(),
		&fakeK8sOperations{},
		&Options{},
	)
	u := &database.User{}
	ctx := context.Background()
	rc, errChan := ops.Create(ctx, "teresa", "test", u, &test.FakeReadSeeker{}, false)

	consumeReader(rc)

	var err error
	select {
	case err = <-errChan:
	default:
	}

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCreateWithRunApp(t *testing.T) {
	fakeExec := exec.NewFakeOperations()
	fakeExec.ExpectedErr = nil
	fakeK8s := &fakeK8sOperations{}

	ops := NewBuildOperations(
		storage.NewFake(),
		app.NewFakeOperations(),
		fakeExec,
		fakeK8s,
		&Options{},
	)
	u := &database.User{}
	ctx := context.Background()
	rc, errChan := ops.Create(ctx, "teresa", "test", u, &test.FakeReadSeeker{}, true)

	consumeReader(rc)

	var err error
	select {
	case err = <-errChan:
	default:
	}

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if !fakeK8s.createServiceWasCalled {
		t.Error("expected `CreateService` was called, but doesn't")
	}
}

func TestCreateWithRunAppRunError(t *testing.T) {
	fakeExec := exec.NewFakeOperations()
	fakeExec.ExpectedErr = fmt.Errorf("some error")
	fakeK8s := &fakeK8sOperations{}

	ops := NewBuildOperations(
		storage.NewFake(),
		app.NewFakeOperations(),
		fakeExec,
		fakeK8s,
		&Options{},
	)
	u := &database.User{}
	ctx := context.Background()
	rc, errChan := ops.Create(ctx, "teresa", "test", u, &test.FakeReadSeeker{}, true)

	consumeReader(rc)

	var err error
	select {
	case err = <-errChan:
	default:
	}

	if err != ErrBuildFail {
		t.Errorf("expected ErrBuildFail, got %v", err)
	}
}
