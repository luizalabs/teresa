package user

import (
	context "golang.org/x/net/context"

	"testing"

	"github.com/luizalabs/teresa-api/pkg/common"
	userpb "github.com/luizalabs/teresa-api/pkg/protobuf"
)

func TestUserLoginSuccess(t *testing.T) {
	fake := NewFakeOperations()

	expectedEmail := "teresa@luizalabs.com"
	expectedPassword := "123456"
	fake.(*FakeOperations).Storage[expectedEmail] = expectedPassword

	s := NewServer(fake)
	r, err := s.Login(
		context.Background(),
		&userpb.LoginRequest{Email: expectedEmail, Password: expectedPassword},
	)
	if err != nil {
		t.Fatal("Got error on make login: ", err)
	}
	if r.Token != "good token" {
		t.Errorf("Expected good token, got %s", r.Token)
	}
}

func TestUserLoginFail(t *testing.T) {
	fake := NewFakeOperations()
	s := NewServer(fake)
	_, err := s.Login(
		context.Background(),
		&userpb.LoginRequest{Email: "invalid@luizalabs.com", Password: "123"},
	)
	if err != common.ErrPermissionDenied {
		t.Error("expected ErrPermisionDenied, got ", err)
	}
}
