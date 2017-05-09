package secrets

import (
	"crypto/rand"
	"crypto/rsa"
)

type Fake struct {
	private *rsa.PrivateKey
	public  *rsa.PublicKey
}

func (f *Fake) PrivateKey() (*rsa.PrivateKey, error) {
	return f.private, nil
}

func (f *Fake) PublicKey() (*rsa.PublicKey, error) {
	return f.public, nil
}

func NewFake() (Secrets, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}
	return &Fake{private: key, public: &key.PublicKey}, nil
}
