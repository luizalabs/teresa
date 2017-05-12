package secrets

import (
	"crypto/rsa"
	"io/ioutil"

	jwt "github.com/dgrijalva/jwt-go"
)

type FileSystemSecretsConfig struct {
	PrivateKey string `split_words:"true" default:"teresa.rsa"`
	PublicKey  string `split_words:"true" default:"teresa.rsa.pub"`
}

type FileSystemSecrets struct {
	privateKey     *rsa.PrivateKey
	publicKey      *rsa.PublicKey
	privateKeyPath string
	publicKeypath  string
}

func (f *FileSystemSecrets) PrivateKey() (*rsa.PrivateKey, error) {
	if f.privateKey != nil {
		return f.privateKey, nil
	}
	b, err := ioutil.ReadFile(f.privateKeyPath)
	if err != nil {
		return nil, err
	}
	f.privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(b)
	return f.privateKey, err
}

func (f *FileSystemSecrets) PublicKey() (*rsa.PublicKey, error) {
	if f.publicKey != nil {
		return f.publicKey, nil
	}
	b, err := ioutil.ReadFile(f.publicKeypath)
	if err != nil {
		return nil, err
	}
	f.publicKey, err = jwt.ParseRSAPublicKeyFromPEM(b)
	return f.publicKey, err
}

func NewFileSystemSecrets(conf FileSystemSecretsConfig) (Secrets, error) {
	s := &FileSystemSecrets{privateKeyPath: conf.PrivateKey, publicKeypath: conf.PublicKey}
	return s, nil
}
