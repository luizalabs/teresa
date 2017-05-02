package secrets

import "testing"

func TestFakeSecret(t *testing.T) {
	f, err := NewFake()
	if err != nil {
		t.Fatal("error on create fake secret: ", err)
	}
	if pk, err := f.PrivateKey(); err != nil || pk == nil {
		t.Errorf("invalid private key generation, key %v, error %v", pk, err)
	}
	if pub, err := f.PublicKey(); err != nil || pub == nil {
		t.Errorf("invalid public key generation, key %v, error %v", pub, err)
	}
}
