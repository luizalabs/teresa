package cloudprovider

import (
	"testing"

	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
	"github.com/pkg/errors"
)

func TestGCECreateOrUpdateSSLSuccess(t *testing.T) {
	ops := &gceOperations{&FakeK8sOperations{HasIngressValue: true}}

	if err := ops.CreateOrUpdateSSL("teresa", "cert", 443); err != nil {
		t.Errorf("got %v; want no error", err)
	}
}

func TestGCECreateOrUpdateSSLFail(t *testing.T) {
	k8s := &FakeK8sOperations{SetIngressAnnotationsErr: errors.New("test"), HasIngressValue: true}
	ops := &gceOperations{k8s}

	e := teresa_errors.ErrInternalServerError
	if err := ops.CreateOrUpdateSSL("teresa", "cert", 443); teresa_errors.Get(err) != e {
		t.Errorf("got %v; want %v", teresa_errors.Get(err), e)
	}
}

func TestGCESSLInfoSuccess(t *testing.T) {
	ops := &gceOperations{&FakeK8sOperations{}}

	if _, err := ops.SSLInfo("teresa"); err != nil {
		t.Errorf("got %v; want no error", err)
	}
}

func TestGCESSLInfoFail(t *testing.T) {
	k8s := &FakeK8sOperations{IngressAnnotationsErr: errors.New("test")}
	ops := &gceOperations{k8s}

	e := teresa_errors.ErrInternalServerError
	if _, err := ops.SSLInfo("teresa"); teresa_errors.Get(err) != e {
		t.Errorf("got %v; want %v", teresa_errors.Get(err), e)
	}
}

func TestGCECreateOrUpdateSSLErrNotImplemented(t *testing.T) {
	ops := &gceOperations{&FakeK8sOperations{HasIngressValue: false}}

	if err := ops.CreateOrUpdateSSL("teresa", "cert", 443); err != ErrNotImplementedOnLoadBalancer {
		t.Errorf("got %v; want %v", err, ErrNotImplementedOnLoadBalancer)
	}
}

func TestGCECreateOrUpdateSSLIngressErr(t *testing.T) {
	want := errors.New("test")
	ops := &gceOperations{&FakeK8sOperations{HasIngressErr: want}}

	if err := ops.CreateOrUpdateSSL("teresa", "cert", 443); err != want {
		t.Errorf("got %v; want %v", err, want)
	}
}

func TestGCECreateOrUpdateStaticIpSuccess(t *testing.T) {
	ops := &gceOperations{&FakeK8sOperations{HasIngressValue: true}}

	if err := ops.CreateOrUpdateStaticIp("teresa", "address-name"); err != nil {
		t.Errorf("got %v; want no error", err)
	}
}

func TestGCECreateOrUpdateStaticIpFail(t *testing.T) {
	k8s := &FakeK8sOperations{SetIngressAnnotationsErr: errors.New("test"), HasIngressValue: true}
	ops := &gceOperations{k8s}

	e := teresa_errors.ErrInternalServerError
	if err := ops.CreateOrUpdateStaticIp("teresa", "address-name"); teresa_errors.Get(err) != e {
		t.Errorf("got %v; want %v", teresa_errors.Get(err), e)
	}
}

func TestGCECreateOrUpdateStaticIpErrNotImplemented(t *testing.T) {
	ops := &gceOperations{&FakeK8sOperations{HasIngressValue: false}}

	if err := ops.CreateOrUpdateStaticIp("teresa", "address-name"); err != ErrNotImplementedOnLoadBalancer {
		t.Errorf("got %v; want %v", err, ErrNotImplementedOnLoadBalancer)
	}
}

func TestGCECreateOrUpdateStaticIpIngressErr(t *testing.T) {
	want := errors.New("test")
	ops := &gceOperations{&FakeK8sOperations{HasIngressErr: want}}

	if err := ops.CreateOrUpdateStaticIp("teresa", "address-name"); err != want {
		t.Errorf("got %v; want %v", err, want)
	}
}
