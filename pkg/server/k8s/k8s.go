package k8s

import (
	"github.com/luizalabs/teresa-api/pkg/server/app"
	"github.com/luizalabs/teresa-api/pkg/server/deploy"
	"k8s.io/client-go/pkg/api"
)

var validServiceTypes = map[api.ServiceType]bool{
	api.ServiceTypeLoadBalancer: true,
	api.ServiceTypeNodePort:     true,
	api.ServiceTypeClusterIP:    true,
}

type Config struct {
	InCluster          bool `default:"true"`
	Host               string
	Username           string
	Password           string
	Insecure           bool   `default:"false"`
	DefaultServiceType string `default:"LoadBalancer"`
}

type Client interface {
	app.K8sOperations
	deploy.K8sOperations
}

func validateConfig(conf *Config) error {
	serviceType := api.ServiceType(conf.DefaultServiceType)
	if _, ok := validServiceTypes[serviceType]; !ok {
		return ErrInvalidServiceType
	}
	return nil
}

func New(conf *Config) (Client, error) {
	if err := validateConfig(conf); err != nil {
		return nil, err
	}
	if conf.InCluster {
		return newInClusterK8sClient(conf)
	}
	return newOutOfClusterK8sClient(conf)
}
