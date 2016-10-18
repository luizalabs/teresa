package k8s

import (
	log "github.com/Sirupsen/logrus"
	"github.com/kelseyhightower/envconfig"
	"k8s.io/kubernetes/pkg/client/restclient"
	"k8s.io/kubernetes/pkg/client/unversioned"
)

// Client is used to keep only one connection with kubernetes
var Client *k8sHelper

// K8sClient is used to keep back compatibility inside the Api
var K8sClient *unversioned.Client // FIXME: this is used just to keep back compatibility... whe should remove ASAP

// K8sConfig struct to accommodate the k8s env config
type k8sConfig struct {
	Host     string `required:"true"`
	Username string `required:"true"`
	Password string `required:"true"`
	Insecure bool   `default:"false"`
}

type k8sHelper struct {
	k8sClient *unversioned.Client
}

func (k *k8sHelper) Apps() AppInterface {
	return newApps(k)
}

func (k *k8sHelper) Users() UserInterface {
	return newUsers(k)
}

func (k *k8sHelper) Teams() TeamInterface {
	return newTeams(k)
}

func (k *k8sHelper) Deployments() DeploymentInterface {
	return newDeployments(k)
}

func (k *k8sHelper) Networks() NetworkInterface {
	return newNetworks(k)
}

func init() {
	// Loading kubernetes (k8s) config from env
	configenv := &k8sConfig{}
	err := envconfig.Process("teresak8s", configenv)
	if err != nil {
		log.Panicf("Failed to read k8s configuration from environment: %s", err.Error())
	}
	// K8s config
	config := &restclient.Config{
		Host:     configenv.Host,
		Username: configenv.Username,
		Password: configenv.Password,
		Insecure: configenv.Insecure,
	}
	// Creating k8s client
	k8sc, err := unversioned.New(config)
	if err != nil {
		log.Panicf("Error trying to create a kubernetes client. Error: %s", err.Error())
	}
	// Exporting k8sHelper with the name Client
	Client = &k8sHelper{k8sClient: k8sc}
	K8sClient = k8sc
}
