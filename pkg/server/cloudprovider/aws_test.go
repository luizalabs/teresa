package cloudprovider

import (
	"testing"

	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
	"github.com/pkg/errors"
)

func TestCreateOrUpdateSSLSuccess(t *testing.T) {
	ops := &awsOperations{&FakeK8sOperations{}}

	if err := ops.CreateOrUpdateSSL("teresa", "cert", 443); err != nil {
		t.Errorf("got %v; want no error", err)
	}
}

func TestCreateOrUpdateSSLFail(t *testing.T) {
	k8s := &FakeK8sOperations{SetServiceAnnotationsErr: errors.New("test")}
	ops := &awsOperations{k8s}

	e := teresa_errors.ErrInternalServerError
	if err := ops.CreateOrUpdateSSL("teresa", "cert", 443); teresa_errors.Get(err) != e {
		t.Errorf("got %v; want %v", teresa_errors.Get(err), e)
	}
}

func TestSSLInfoSuccess(t *testing.T) {
	ops := &awsOperations{&FakeK8sOperations{}}

	if _, err := ops.SSLInfo("teresa"); err != nil {
		t.Errorf("got %v; want no error", err)
	}
}

func TestSSLInfoFail(t *testing.T) {
	k8s := &FakeK8sOperations{ServiceAnnotationsErr: errors.New("test")}
	ops := &awsOperations{k8s}

	e := teresa_errors.ErrInternalServerError
	if _, err := ops.SSLInfo("teresa"); teresa_errors.Get(err) != e {
		t.Errorf("got %v; want %v", teresa_errors.Get(err), e)
	}
}

func TestCreateOrUpdateSSLErrNotImplemented(t *testing.T) {
	ops := &awsOperations{&FakeK8sOperations{HasIngressValue: true}}

	if err := ops.CreateOrUpdateSSL("teresa", "cert", 443); err != ErrNotImplementedOnIngress {
		t.Errorf("got %v; want %v", err, ErrNotImplementedOnIngress)
	}
}

func TestCreateOrUpdateSSLIngressErr(t *testing.T) {
	want := errors.New("test")
	ops := &awsOperations{&FakeK8sOperations{HasIngressErr: want}}

	if err := ops.CreateOrUpdateSSL("teresa", "cert", 443); err != want {
		t.Errorf("got %v; want %v", err, want)
	}
}
