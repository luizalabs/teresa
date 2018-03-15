package service

import (
	"github.com/luizalabs/teresa/pkg/server/database"
)

type FakeOperations struct {
	EnableSSLErr error
	InfoErr      error
	InfoValue    *Info
}

type FakeCloudProviderOperations struct {
	CreateOrUpdateSSLErr error
	SSLInfoErr           error
	SSLInfoValue         *SSLInfo
}

type FakeAppOperations struct {
	NegateHasPermission bool
}

type FakeK8sOperations struct {
	UpdateServicePortsErr error
	IsNotFoundErr         bool
	ServicePortsErr       error
	ServicePortsValue     []*ServicePort
}

func (f *FakeOperations) EnableSSL(user *database.User, appName, cert string, only bool) error {
	return f.EnableSSLErr
}

func (f *FakeOperations) Info(user *database.User, appName string) (*Info, error) {
	return f.InfoValue, f.InfoErr
}

func (f *FakeCloudProviderOperations) CreateOrUpdateSSL(appName, cert string, port int) error {
	return f.CreateOrUpdateSSLErr
}

func (f *FakeCloudProviderOperations) SSLInfo(appName string) (*SSLInfo, error) {
	return f.SSLInfoValue, f.SSLInfoErr
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

func (f *FakeK8sOperations) ServicePorts(namespace, svcName string) ([]*ServicePort, error) {
	return f.ServicePortsValue, f.ServicePortsErr
}
