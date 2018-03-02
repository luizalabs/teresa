package k8s

import (
	"time"

	"k8s.io/client-go/pkg/api"
)

var validServiceTypes = map[api.ServiceType]bool{
	api.ServiceTypeLoadBalancer: true,
	api.ServiceTypeNodePort:     true,
	api.ServiceTypeClusterIP:    true,
}

type Config struct {
	ConfigFile    string        `split_words:"true"`
	PodRunTimeout time.Duration `split_words:"true" default:"30m"`
	Ingress       bool          `split_words:"true" default:"false"`
}

func New(conf *Config) (*Client, error) {
	if conf.ConfigFile == "" {
		return newInClusterK8sClient(conf)
	}
	return newOutOfClusterK8sClient(conf)
}
