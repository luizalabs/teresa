package user

import (
	context "golang.org/x/net/context"

	"testing"

	"github.com/luizalabs/teresa-api/models/storage"
	userpb "github.com/luizalabs/teresa-api/pkg/protobuf"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

func TestUserLoginSuccess(t *testing.T) {
	fake := NewFakeOperations()

	expectedEmail := "teresa@luizalabs.com"
	expectedPassword := "123456"
	fake.(*FakeOperations).Storage[expectedEmail] = expectedPassword

	s := NewService(fake)
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
	s := NewService(fake)
	_, err := s.Login(
		context.Background(),
		&userpb.LoginRequest{Email: "invalid@luizalabs.com", Password: "123"},
	)
	if err != auth.ErrPermissionDenied {
		t.Error("expected ErrPermisionDenied, got ", err)
	}
}

func TestSetPasswordSuccess(t *testing.T) {
	fake := NewFakeOperations()

	expectedEmail := "teresa@luizalabs.com"
	expectedPassword := "123456"
	fake.(*FakeOperations).Storage[expectedEmail] = "gopher"

	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &storage.User{Email: expectedEmail})
	_, err := s.SetPassword(
		ctx,
		&userpb.SetPasswordRequest{Password: expectedPassword},
	)
	if err != nil {
		t.Error("Got error on make SetPassword: ", err)
	}
}
