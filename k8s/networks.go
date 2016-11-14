package k8s

import (
	"github.com/luizalabs/teresa-api/models"
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

func newNetworks(c *k8sHelper) *networks {
	return &networks{k: c}
}

func (c networks) CreateLoadBalancerService(app *models.App) error {
	srv := newService(app, api.ServiceTypeLoadBalancer)
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
