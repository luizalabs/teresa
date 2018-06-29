package service

import (
	"errors"
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
)

func setupTestOps() *ServiceOperations {
	fakeAppOps := &FakeAppOperations{App: &app.App{}}
	fakeCloudProviderOps := &FakeCloudProviderOperations{}
	fakeK8sOps := &FakeK8sOperations{ServiceValue: &spec.Service{}}
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
	ops.k8s.(*FakeK8sOperations).ServiceErr = errors.New("test")
	ops.k8s.(*FakeK8sOperations).IsNotFoundErr = true
	user := &database.User{}

	if _, err := ops.Info(user, "teresa"); err != ErrNotFound {
		t.Errorf("got %v; want %v", err, ErrNotFound)
	}
}

func TestOpsInfoServicePortsFail(t *testing.T) {
	ops := setupTestOps()
	ops.k8s.(*FakeK8sOperations).ServiceErr = errors.New("test")
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

func TestOpsWhitelistSourceRangesSuccess(t *testing.T) {
	ops := setupTestOps()
	user := &database.User{}
	ranges := []string{"range1", "range2"}

	if err := ops.WhitelistSourceRanges(user, "teresa", ranges); err != nil {
		t.Error("got unexpected error:", err)
	}
}

func TestOpsWhitelistSourceRangesAppNotFound(t *testing.T) {
	ops := setupTestOps()
	ops.aops.(*FakeAppOperations).CheckPermAndGetErr = app.ErrNotFound
	user := &database.User{}
	ranges := []string{"range1", "range2"}

	if err := ops.WhitelistSourceRanges(user, "test", ranges); err != app.ErrNotFound {
		t.Errorf("got %v; want %v", err, app.ErrNotFound)
	}
}

func TestOpsWhitelistSourceRangesPermissionDenied(t *testing.T) {
	ops := setupTestOps()
	ops.aops.(*FakeAppOperations).NegateHasPermission = true
	user := &database.User{}
	ranges := []string{"range1", "range2"}

	if err := ops.WhitelistSourceRanges(user, "teresa", ranges); err != auth.ErrPermissionDenied {
		t.Errorf("got %v; want %v", err, auth.ErrPermissionDenied)
	}
}

func TestOpsWhitelistSourceRangesServiceNotFound(t *testing.T) {
	ops := setupTestOps()
	ops.k8s.(*FakeK8sOperations).SetLoadBalancerSourceRangesErr = errors.New("test")
	ops.k8s.(*FakeK8sOperations).IsNotFoundErr = true
	user := &database.User{}
	ranges := []string{"range1", "range2"}

	if err := ops.WhitelistSourceRanges(user, "teresa", ranges); err != ErrNotFound {
		t.Errorf("got %v; want %v", err, ErrNotFound)
	}
}

func TestOpsWhitelistSourceRangesInvalid(t *testing.T) {
	ops := setupTestOps()
	ops.k8s.(*FakeK8sOperations).SetLoadBalancerSourceRangesErr = errors.New("test")
	ops.k8s.(*FakeK8sOperations).IsInvalidErr = true
	user := &database.User{}
	ranges := []string{"range1", "range2"}

	if err := ops.WhitelistSourceRanges(user, "teresa", ranges); err != ErrInvalidSourceRanges {
		t.Errorf("got %v; want %v", err, ErrInvalidSourceRanges)
	}
}

func TestOpsWhitelistSourceRangesNotImplemented(t *testing.T) {
	ops := setupTestOps()
	ops.aops.(*FakeAppOperations).App.Internal = true
	user := &database.User{}
	ranges := []string{"range1", "range2"}

	if err := ops.WhitelistSourceRanges(user, "teresa", ranges); err != ErrWhitelistUnimplemented {
		t.Errorf("got %v; want %v", err, ErrWhitelistUnimplemented)
	}
}

func TestOpsWhitelistSourceRangesInternalServerError(t *testing.T) {
	ops := setupTestOps()
	ops.k8s.(*FakeK8sOperations).SetLoadBalancerSourceRangesErr = errors.New("test")
	user := &database.User{}
	ranges := []string{"range1", "range2"}

	e := teresa_errors.ErrInternalServerError
	if err := ops.WhitelistSourceRanges(user, "teresa", ranges); teresa_errors.Get(err) != e {
		t.Errorf("got %v; want %v", teresa_errors.Get(err), e)
	}
}
