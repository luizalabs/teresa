package secrets

import "crypto/rsa"

type Secrets interface {
	PrivateKey() (*rsa.PrivateKey, error)
	PublicKey() (*rsa.PublicKey, error)
}
