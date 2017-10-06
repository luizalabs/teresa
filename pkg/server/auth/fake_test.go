package auth

import (
	"testing"
	"time"
)

func TestFakeGenerateToken(t *testing.T) {
	fake := NewFake()
	token, err := fake.GenerateToken("gopher@luizalabs.com", time.Second)
	if err != nil {
		t.Fatal("error on generate fake token: ", err)
	}
	if token != "good token" {
		t.Error("expected 'good token', got: ", token)
	}
}

func TestFakeValidateToken(t *testing.T) {
	fake := NewFake()
	email, err := fake.ValidateToken("foo")
	if err != nil {
		t.Fatal("error on validate a fake token: ", err)
	}
	if email != "gopher@luizalabs.com" {
		t.Errorf("expected gopher@luizalabs.com, got %s", email)
	}
}
