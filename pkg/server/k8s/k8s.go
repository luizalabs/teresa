package k8s

import (
	"time"

	k8sv1 "k8s.io/api/core/v1"
)

var validServiceTypes = map[k8sv1.ServiceType]bool{
	k8sv1.ServiceTypeLoadBalancer: true,
	k8sv1.ServiceTypeNodePort:     true,
	k8sv1.ServiceTypeClusterIP:    true,
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
