package service

import (
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
)

const (
	defaultPortName    = "tcp"
	defaultSSLPortName = "ssl"
	sslPort            = 443
)

type ServicePort struct {
	Name       string
	Port       int
	TargetPort int
}

type CloudProviderOperations interface {
	CreateOrUpdateSSL(appName, cert string, port int) error
}

type K8sOperations interface {
	UpdateServicePorts(namespace, svcName string, ports []ServicePort) error
	IsNotFound(err error) bool
}

type AppOperations interface {
	HasPermission(user *database.User, appName string) bool
}

type Operations interface {
	EnableSSL(user *database.User, appName, cert string, only bool) error
}

type ServiceOperations struct {
	aops AppOperations
	cops CloudProviderOperations
	k8s  K8sOperations
}

func (ops *ServiceOperations) EnableSSL(user *database.User, appName, cert string, only bool) error {
	if !ops.aops.HasPermission(user, appName) {
		return auth.ErrPermissionDenied
	}
	if err := ops.cops.CreateOrUpdateSSL(appName, cert, sslPort); err != nil {
		return err
	}
	ports := []ServicePort{
		{Name: defaultPortName, TargetPort: spec.DefaultPort},
		{Name: defaultSSLPortName, Port: sslPort, TargetPort: spec.DefaultPort},
	}
	if only {
		ports = ports[1:]
	}
	if err := ops.k8s.UpdateServicePorts(appName, appName, ports); err != nil {
		if ops.k8s.IsNotFound(err) {
			return ErrNotFound
		}
		return teresa_errors.NewInternalServerError(err)
	}
	return nil
}

func NewOperations(aops AppOperations, cops CloudProviderOperations, k8s K8sOperations) *ServiceOperations {
	return &ServiceOperations{aops: aops, cops: cops, k8s: k8s}
}
