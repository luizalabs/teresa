package service

import (
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/spec"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
)

const (
	defaultSSLPortName = "ssl"
	sslPort            = 443
)

type CloudProviderOperations interface {
	CreateOrUpdateSSL(appName, cert string, port int) error
	CreateOrUpdateStaticIp(appName, addressName string) error
	SSLInfo(appName string) (*SSLInfo, error)
}

type K8sOperations interface {
	UpdateServicePorts(namespace, svcName string, ports []spec.ServicePort) error
	IsNotFound(err error) bool
	IsInvalid(err error) bool
	Service(namespace, svcName string) (*spec.Service, error)
	SetLoadBalancerSourceRanges(namespace, svcName string, sourceRanges []string) error
	HasIngress(namespace, name string) (bool, error)
}

type AppOperations interface {
	HasPermission(user *database.User, appName string) bool
	CheckPermAndGet(user *database.User, appName string) (*app.App, error)
}

type Operations interface {
	EnableSSL(user *database.User, appName, cert string, only bool) error
	SetStaticIp(user *database.User, appName, addressName string) error
	Info(user *database.User, appName string) (*Info, error)
	WhitelistSourceRanges(user *database.User, appName string, sourceRanges []string) error
}

type ServiceOperations struct {
	aops AppOperations
	cops CloudProviderOperations
	k8s  K8sOperations
}

func (ops *ServiceOperations) EnableSSL(user *database.User, appName, cert string, only bool) error {
	app, err := ops.aops.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}
	if err := ops.cops.CreateOrUpdateSSL(appName, cert, sslPort); err != nil {
		return err
	}

	hasIngress, err := ops.k8s.HasIngress(app.Name, app.Name)

	if !hasIngress {
		ports := []spec.ServicePort{
			*spec.NewDefaultServicePort(app.Protocol),
			*spec.NewServicePort(defaultSSLPortName, sslPort, spec.DefaultPort),
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
	}
	return nil
}

func (ops *ServiceOperations) SetStaticIp(user *database.User, appName, addressName string) error {
	_, err := ops.aops.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}
	return ops.cops.CreateOrUpdateStaticIp(appName, addressName)
}

func (ops *ServiceOperations) Info(user *database.User, appName string) (*Info, error) {
	if !ops.aops.HasPermission(user, appName) {
		return nil, auth.ErrPermissionDenied
	}
	ssl, err := ops.cops.SSLInfo(appName)
	if err != nil {
		return nil, err
	}
	svc, err := ops.k8s.Service(appName, appName)
	if err != nil {
		if ops.k8s.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, teresa_errors.NewInternalServerError(err)
	}
	info := &Info{
		SSLInfo:      ssl,
		ServicePorts: svc.Ports,
		SourceRanges: svc.SourceRanges,
	}
	return info, nil
}

func (ops *ServiceOperations) WhitelistSourceRanges(user *database.User, appName string, sourceRanges []string) error {
	a, err := ops.aops.CheckPermAndGet(user, appName)
	if err != nil {
		return err
	}
	hasIngress, err := ops.k8s.HasIngress(a.Name, a.Name)
	if err != nil {
		return err
	}
	if hasIngress || a.Internal {
		return ErrWhitelistUnimplemented
	}
	if err := ops.k8s.SetLoadBalancerSourceRanges(appName, appName, sourceRanges); err != nil {
		if ops.k8s.IsNotFound(err) {
			return ErrNotFound
		} else if ops.k8s.IsInvalid(err) {
			return ErrInvalidSourceRanges
		}
		return teresa_errors.NewInternalServerError(err)
	}
	return nil
}

func NewOperations(aops AppOperations, cops CloudProviderOperations, k8s K8sOperations) *ServiceOperations {
	return &ServiceOperations{aops: aops, cops: cops, k8s: k8s}
}
