package k8s

import (
	"github.com/luizalabs/teresa-api/pkg/server/app"
	"k8s.io/client-go/pkg/api"
)

var validServiceTypes = map[api.ServiceType]bool{
	api.ServiceTypeLoadBalancer: true,
	api.ServiceTypeNodePort:     true,
	api.ServiceTypeClusterIP:    true,
}

type Config struct {
	InCluster          bool   `default:"true"`
	Host               string `required:"true"`
	Username           string `required:"true"`
	Password           string `required:"true"`
	Insecure           bool   `default:"false"`
	DefaultServiceType string `default:"LoadBalancer"`
}

type Client interface {
	app.K8sOperations
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
		return newInClusterK8sClient()
	}
	return newOutOfClusterK8sClient(conf)
}
