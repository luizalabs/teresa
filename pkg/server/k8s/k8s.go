package k8s

import (
	"io"
	"time"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/deploy"
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
	PodRunTimeout      time.Duration `split_words:"true" default:"30m"`
	Ingress            bool          `split_words:"true" default:"false"`
}

type Namespace interface {
	NamespaceAnnotation(namespace, annotation string) (string, error)
	NamespaceLabel(namespace, label string) (string, error)
	CreateNamespace(a *app.App, userEmail string) error
	NamespaceListByLabel(label, value string) ([]string, error)
	DeleteNamespace(namespace string) error
	SetNamespaceLabels(namespace string, labels map[string]string) error
	SetNamespaceAnnotations(namespace string, annotations map[string]string) error
	Status(namespace string) (*app.Status, error)
}

type Pod interface {
	PodList(namespace string) ([]*app.Pod, error)
	PodLogs(namespace, podName string, lines int64, follow bool) (io.ReadCloser, error)
	DeletePod(namespace, podName string) error
	PodRun(podSpec *deploy.PodSpec) (io.ReadCloser, <-chan int, error)
}

type Quota interface {
	CreateQuota(a *app.App) error
	Limits(namespace, name string) (*app.Limits, error)
}

type Secret interface {
	CreateSecret(appName, secretName string, data map[string][]byte) error
}

type Hpa interface {
	Autoscale(namespace string) (*app.Autoscale, error)
	CreateOrUpdateAutoscale(a *app.App) error
}

type Deploy interface {
	DeploySetReplicas(namespace, name string, replicas int32) error
	CreateOrUpdateDeployEnvVars(namespace, name string, evs []*app.EnvVar) error
	DeleteDeployEnvVars(namespace, name string, evNames []string) error
	CreateOrUpdateDeploy(deploySpec *deploy.DeploySpec) error
	ExposeDeploy(namespace, name, vHost string, w io.Writer) error
	ReplicaSetListByLabel(namespace, label, value string) ([]*deploy.ReplicaSetListItem, error)
	DeployRollbackToRevision(namespace, name, revision string) error
}

type Error interface {
	IsNotFound(err error) bool
	IsAlreadyExists(err error) bool
}

type Service interface {
	AddressList(namespace string) ([]*app.Address, error)
}

type TeresaHealthCheck interface {
	HealthCheck() error
}

type Client interface {
	Namespace
	Pod
	Quota
	Secret
	Hpa
	Deploy
	Error
	Service
	TeresaHealthCheck
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
