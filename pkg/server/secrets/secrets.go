package secrets

import (
	"crypto/rsa"
	"crypto/tls"
)

type Secrets interface {
	PrivateKey() (*rsa.PrivateKey, error)
	PublicKey() (*rsa.PublicKey, error)
	TLSCertificate() (*tls.Certificate, error)
}
