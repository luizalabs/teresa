package k8s

import (
	"time"

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
	ConfigFile         string        `split_words:"true"`
	DefaultServiceType string        `split_words:"true" default:"LoadBalancer"`
	Timeout            time.Duration `default:"30s"`
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
	if conf.ConfigFile == "" {
		return newInClusterK8sClient(conf)
	}
	return newOutOfClusterK8sClient(conf)
}
