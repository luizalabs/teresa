package cloudprovider

import "github.com/luizalabs/teresa/pkg/server/service"

type gceOperations struct {
	k8s K8sOperations
}

func (ops *gceOperations) CreateOrUpdateSSL(appName, cert string, port int) error {
	return ErrNotImplemented
}

func (ops *gceOperations) SSLInfo(appName string) (*service.SSLInfo, error) {
	return nil, ErrNotImplemented
}

func (ops *gceOperations) Name() string {
	return "gce"
}
