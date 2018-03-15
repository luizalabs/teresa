package service

import (
	"errors"
	"testing"

	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
)

func setupTestOps() *ServiceOperations {
	fakeAppOps := &FakeAppOperations{}
	fakeCloudProviderOps := &FakeCloudProviderOperations{}
	fakeK8sOps := &FakeK8sOperations{}
	return NewOperations(fakeAppOps, fakeCloudProviderOps, fakeK8sOps)
}

func TestOpsEnableSSLSuccess(t *testing.T) {
	ops := setupTestOps()
	user := &database.User{}
	tc := []bool{true, false}

	for _, only := range tc {
		if err := ops.EnableSSL(user, "teresa", "cert", only); err != nil {
			t.Errorf("got %v; want no error", err)
		}
	}
}

func TestOpsEnableSSLPermissionDenied(t *testing.T) {
	ops := setupTestOps()
	ops.aops.(*FakeAppOperations).NegateHasPermission = true
	user := &database.User{}

	if err := ops.EnableSSL(user, "teresa", "cert", false); err != auth.ErrPermissionDenied {
		t.Errorf("got %v; want %v", err, auth.ErrPermissionDenied)
	}
}

func TestOpsEnableSSLCloudProviderFail(t *testing.T) {
	ops := setupTestOps()
	wantErr := errors.New("test")
	ops.cops.(*FakeCloudProviderOperations).CreateOrUpdateSSLErr = wantErr
	user := &database.User{}

	if err := ops.EnableSSL(user, "teresa", "cert", false); err != wantErr {
		t.Errorf("got %v; want %v", err, wantErr)
	}
}

func TestOpsEnableSSLUpdateServicePortsFail(t *testing.T) {
	ops := setupTestOps()
	ops.k8s.(*FakeK8sOperations).UpdateServicePortsErr = errors.New("test")
	user := &database.User{}

	e := teresa_errors.ErrInternalServerError
	if err := ops.EnableSSL(user, "teresa", "cert", false); teresa_errors.Get(err) != e {
		t.Errorf("got %v; want %v", teresa_errors.Get(err), e)
	}
}

func TestOpsEnableSSLServiceNotFound(t *testing.T) {
	ops := setupTestOps()
	ops.k8s.(*FakeK8sOperations).UpdateServicePortsErr = errors.New("test")
	ops.k8s.(*FakeK8sOperations).IsNotFoundErr = true
	user := &database.User{}

	if err := ops.EnableSSL(user, "teresa", "cert", false); err != ErrNotFound {
		t.Errorf("got %v; want %v", err, ErrNotFound)
	}
}

func TestOpsInfoSuccess(t *testing.T) {
	ops := setupTestOps()
	user := &database.User{}

	if _, err := ops.Info(user, "teresa"); err != nil {
		t.Errorf("got %v; want no error", err)
	}
}

func TestOpsInfoPermissionDenied(t *testing.T) {
	ops := setupTestOps()
	ops.aops.(*FakeAppOperations).NegateHasPermission = true
	user := &database.User{}

	if _, err := ops.Info(user, "teresa"); err != auth.ErrPermissionDenied {
		t.Errorf("got %v; want %v", err, auth.ErrPermissionDenied)
	}
}

func TestOpsInfoServiceNotFound(t *testing.T) {
	ops := setupTestOps()
	ops.k8s.(*FakeK8sOperations).ServicePortsErr = errors.New("test")
	ops.k8s.(*FakeK8sOperations).IsNotFoundErr = true
	user := &database.User{}

	if _, err := ops.Info(user, "teresa"); err != ErrNotFound {
		t.Errorf("got %v; want %v", err, ErrNotFound)
	}
}

func TestOpsInfoServicePortsFail(t *testing.T) {
	ops := setupTestOps()
	ops.k8s.(*FakeK8sOperations).ServicePortsErr = errors.New("test")
	user := &database.User{}

	e := teresa_errors.ErrInternalServerError
	if _, err := ops.Info(user, "teresa"); teresa_errors.Get(err) != e {
		t.Errorf("got %v; want %v", teresa_errors.Get(err), e)
	}
}

func TestOpsInfoCloudProviderFail(t *testing.T) {
	ops := setupTestOps()
	wantErr := errors.New("test")
	ops.cops.(*FakeCloudProviderOperations).SSLInfoErr = wantErr
	user := &database.User{}

	if _, err := ops.Info(user, "teresa"); err != wantErr {
		t.Errorf("got %v; want %v", err, wantErr)
	}
}
