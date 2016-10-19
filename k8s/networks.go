package k8s

import (
	log "github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
	"github.com/luizalabs/tapi/models"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/util/intstr"
)

// NetworksInterface is used to allow mock testing
type NetworksInterface interface {
	Networks() DeploymentInterface
}

// NetworkInterface is used to interact with Kubernetes and also to allow mock testing
type NetworkInterface interface {
	CreateLoadBalancerService(app *models.App) error
	GetService(appName string) (srv *api.Service, err error)
}

type networks struct {
	k *k8sHelper
}

type networksConfig struct {
	DefaultServiceType string `envconfig:"DEFAULT_SERVICE_TYPE" default:"LoadBalancer"`
}

var conf networksConfig

var validServiceTypes = []api.ServiceType{
	api.ServiceTypeLoadBalancer,
	api.ServiceTypeNodePort,
	api.ServiceTypeClusterIP,
}

func newNetworks(c *k8sHelper) *networks {
	return &networks{k: c}
}

func (c networks) CreateLoadBalancerService(app *models.App) error {
	serviceType := api.ServiceType(conf.DefaultServiceType)
	srv := newService(app, serviceType)
	_, err := c.k.k8sClient.Services(*app.Name).Create(srv)
	return err
}

func newService(app *models.App, serviceType api.ServiceType) (srv *api.Service) {
	srv = &api.Service{
		TypeMeta: unversioned.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: api.ObjectMeta{
			Labels: map[string]string{
				"run": *app.Name,
			},
			Name:      *app.Name,
			Namespace: *app.Name,
		},
		Spec: api.ServiceSpec{
			Type:            serviceType,
			SessionAffinity: api.ServiceAffinityNone,
			Selector: map[string]string{
				"run": *app.Name,
			},
			Ports: []api.ServicePort{
				api.ServicePort{
					Port:       80,
					Protocol:   api.ProtocolTCP,
					TargetPort: intstr.FromInt(5000),
				},
			},
		},
	}
	return
}

func (c networks) GetService(name string) (srv *api.Service, err error) {
	srv, err = c.k.k8sClient.Services(name).Get(name)
	return
}

func init() {
	err := envconfig.Process("teresa_network", &conf)
	if err != nil {
		log.Fatalf("failed to read the network configuration from environment: %s", err.Error())
	}
	for _, v := range validServiceTypes {
		if v == api.ServiceType(conf.DefaultServiceType) {
			return
		}
	}
	log.Fatalf("invalid default service type: %s", conf.DefaultServiceType)
}
