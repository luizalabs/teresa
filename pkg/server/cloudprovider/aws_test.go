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
