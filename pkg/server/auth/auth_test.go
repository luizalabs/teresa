package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

func TestJWTAuthGenerateToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal("error on create fake secrets: ", err)
	}
	pub := key.PublicKey

	a := New(key, &pub)
	token, err := a.GenerateToken("gopher@luizalabs.com")
	if err != nil {
		t.Fatal("error on generate token: ", err)
	}
	if token == "" {
		t.Error("expected a valid token, got ", token)
	}
}
