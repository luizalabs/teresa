package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
)

var (
	privateKey, _ = rsa.GenerateKey(rand.Reader, 2048)
	publicKey     = &privateKey.PublicKey
)

func TestJWTAuthGenerateToken(t *testing.T) {
	a := New(privateKey, publicKey)
	token, err := a.GenerateToken("gopher@luizalabs.com")
	if err != nil {
		t.Fatal("error on generate token: ", err)
	}
	if token == "" {
		t.Error("expected a valid token, got ", token)
	}
}

func TestJWTAuthValidateTokenSuccess(t *testing.T) {
	a := New(privateKey, publicKey)

	expectedEmail := "gopher@luizalabs.com"
	token, err := a.GenerateToken(expectedEmail)
	if err != nil {
		t.Fatal("error on generate token: ", err)
	}

	email, err := a.ValidateToken(token)
	if err != nil {
		t.Fatal("error on validate token: ", err)
	}

	if email != expectedEmail {
		t.Errorf("expected %s, got %s", expectedEmail, email)
	}
}

func TestJWTAuthValidateTokenForInvalidToken(t *testing.T) {
	a := New(privateKey, publicKey)
	if _, err := a.ValidateToken("invalid@foo.com"); err != ErrPermissionDenied {
		t.Error("expected ErrPermissionDenied, got nil")
	}
}
