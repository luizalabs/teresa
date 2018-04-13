package service

import (
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/spec"
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
	App                 *app.App
}

type FakeK8sOperations struct {
	UpdateServicePortsErr error
	IsNotFoundErr         bool
	ServicePortsErr       error
	ServicePortsValue     []*spec.ServicePort
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

func (f *FakeAppOperations) CheckPermAndGet(user *database.User, appName string) (*app.App, error) {
	var err error
	if f.NegateHasPermission {
		err = auth.ErrPermissionDenied
	}
	return f.App, err
}

func (f *FakeK8sOperations) UpdateServicePorts(namespace, svcName string, ports []spec.ServicePort) error {
	return f.UpdateServicePortsErr
}

func (f *FakeK8sOperations) IsNotFound(err error) bool {
	return f.IsNotFoundErr
}

func (f *FakeK8sOperations) ServicePorts(namespace, svcName string) ([]*spec.ServicePort, error) {
	return f.ServicePortsValue, f.ServicePortsErr
}
