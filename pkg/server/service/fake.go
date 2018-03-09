package service

import (
	"github.com/luizalabs/teresa/pkg/server/database"
)

type FakeOperations struct {
	EnableSSLErr error
}

type FakeCloudProviderOperations struct {
	CreateOrUpdateSSLErr error
}

type FakeAppOperations struct {
	NegateHasPermission bool
}

type FakeK8sOperations struct {
	UpdateServicePortsErr error
	IsNotFoundErr         bool
}

func (f *FakeOperations) EnableSSL(user *database.User, appName, cert string, only bool) error {
	return f.EnableSSLErr
}

func (f *FakeCloudProviderOperations) CreateOrUpdateSSL(appName, cert string, port int) error {
	return f.CreateOrUpdateSSLErr
}

func (f *FakeAppOperations) HasPermission(user *database.User, appName string) bool {
	return !f.NegateHasPermission
}

func (f *FakeK8sOperations) UpdateServicePorts(namespace, svcName string, ports []ServicePort) error {
	return f.UpdateServicePortsErr
}

func (f *FakeK8sOperations) IsNotFound(err error) bool {
	return f.IsNotFoundErr
}
