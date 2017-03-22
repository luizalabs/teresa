package k8s

import (
	"k8s.io/kubernetes/pkg/api"
)

// SecretInterface is used to allow mock testing
type SecretsInterface interface {
	Secrets() SecretInterface
}

// SecretInterface is used to interact with Kubernetes and also to allow mock testing
type SecretInterface interface {
	Get(appName, name string) (secret *api.Secret, err error)
}

type secrets struct {
	k *k8sHelper
}

func newSecrets(k *k8sHelper) *secrets {
	return &secrets{k: k}
}

func (s *secrets) Get(appName, name string) (*api.Secret, error) {
	return s.k.k8sClient.Secrets(appName).Get(name)
}
