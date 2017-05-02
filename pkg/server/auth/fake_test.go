package auth

import "testing"

func TestFakeGenerateToken(t *testing.T) {
	fake := NewFake()
	token, err := fake.GenerateToken("gopher@luizalabs.com")
	if err != nil {
		t.Fatal("error on generate fake token: ", err)
	}
	if token != "good token" {
		t.Error("expected 'good token', got: ", token)
	}
}
