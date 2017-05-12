package secrets

import (
	"path/filepath"
	"testing"
)

func TestFileSystemSecrets(t *testing.T) {
	f, err := NewFileSystemSecrets(FileSystemSecretsConfig{
		PrivateKey: filepath.Join("testdata", "fake.rsa"),
		PublicKey:  filepath.Join("testdata", "fake.rsa.pub"),
	})
	if err != nil {
		t.Fatal("error on create file system secret: ", err)
	}
	if pk, err := f.PrivateKey(); err != nil || pk == nil {
		t.Errorf("invalid private key generation, key %v, error %v", pk, err)
	}
	if pub, err := f.PublicKey(); err != nil || pub == nil {
		t.Errorf("invalid public key generation, key %v, error %v", pub, err)
	}
}
