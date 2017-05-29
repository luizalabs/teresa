package k8s

import (
	"k8s.io/kubernetes/pkg/api"
)

var validServiceTypes = map[api.ServiceType]bool{
	api.ServiceTypeLoadBalancer: true,
	api.ServiceTypeNodePort:     true,
	api.ServiceTypeClusterIP:    true,
}

type Config struct {
	Host               string `required:"true"`
	Username           string `required:"true"`
	Password           string `required:"true"`
	Insecure           bool   `default:"false"`
	DefaultServiceType string `default:"LoadBalancer"`
}

type Client interface{}

func New(conf *Config) (Client, error) {
	serviceType := api.ServiceType(conf.DefaultServiceType)
	if _, ok := validServiceTypes[serviceType]; !ok {
		return nil, ErrInvalidServiceType
	}
	return newUnversioned(conf)
}
