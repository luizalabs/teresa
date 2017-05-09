package secrets

import (
	"crypto/rsa"
	"errors"
	"os"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/luizalabs/teresa-api/k8s"
)

const (
	privateKeyName = "teresa.rsa"
	publicKeyName  = "teresa.rsa.pub"
	k8sSecretName  = "teresa-keys"
)

var (
	ErrMissingKey          = errors.New("Missing auth key")
	ErrMissingNamespaceEnv = errors.New("Missing 'NAMESPACE' environment variable")
)

type Secrets interface {
	PrivateKey() (*rsa.PrivateKey, error)
	PublicKey() (*rsa.PublicKey, error)
}

type K8SSecrets struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	namespace  string
}

func (ks *K8SSecrets) PrivateKey() (*rsa.PrivateKey, error) {
	if ks.privateKey != nil {
		return ks.privateKey, nil
	}
	b, err := ks.getKeyBytes(privateKeyName)
	if err != nil {
		return nil, err
	}
	ks.privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(b)
	return ks.privateKey, err
}

func (ks *K8SSecrets) PublicKey() (*rsa.PublicKey, error) {
	if ks.publicKey != nil {
		return ks.publicKey, nil
	}
	b, err := ks.getKeyBytes(publicKeyName)
	if err != nil {
		return nil, err
	}
	ks.publicKey, err = jwt.ParseRSAPublicKeyFromPEM(b)
	return ks.publicKey, err
}

func (ks *K8SSecrets) getKeyBytes(keyName string) ([]byte, error) {
	s, err := k8s.Client.Secrets().Get(ks.namespace, k8sSecretName)
	if err != nil {
		return nil, err
	}
	data, ok := s.Data[keyName]
	if !ok {
		return nil, ErrMissingKey
	}
	return data, nil
}

func NewK8SSecrets() (Secrets, error) {
	ns, ok := os.LookupEnv("NAMESPACE")
	if !ok {
		return nil, ErrMissingNamespaceEnv
	}
	return &K8SSecrets{namespace: ns}, nil
}
