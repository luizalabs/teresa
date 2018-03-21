package cloudprovider

import "github.com/luizalabs/teresa/pkg/server/service"

type fallbackOperations struct{}

func (*fallbackOperations) SSLInfo(appName string) (*service.SSLInfo, error) {
	return nil, ErrNotImplemented
}

func (*fallbackOperations) CreateOrUpdateSSL(appName, cert string, port int) error {
	return ErrNotImplemented
}
