package cloudprovider

import (
	"github.com/luizalabs/teresa/pkg/server/service"
)

type Operations interface {
	CreateOrUpdateSSL(appName, cert string, port int) error
	SSLInfo(appName string) (*service.SSLInfo, error)
	Name() string
}

type K8sOperations interface {
	CloudProviderName() (string, error)
	SetServiceAnnotations(namespace, service string, annotations map[string]string) error
	ServiceAnnotations(namespace, service string) (map[string]string, error)
	IsNotFound(err error) bool
	HasIngress(namespace, name string) (bool, error)
}

func NewOperations(k8s K8sOperations) Operations {
	name, err := k8s.CloudProviderName()
	if err != nil {
		return &fallbackOperations{}
	}
	switch name {
	case "aws":
		return &awsOperations{k8s: k8s}
	case "gce":
		return &gceOperations{k8s: k8s}
	default:
		return &fallbackOperations{}
	}
}
