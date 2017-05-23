package app

import (
	context "golang.org/x/net/context"

	"testing"

	"github.com/luizalabs/teresa-api/models/storage"
	appb "github.com/luizalabs/teresa-api/pkg/protobuf/app"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

func TestCreateSuccess(t *testing.T) {
	fake := NewFakeOperations()
	user := &storage.User{Email: "gopher@luizalabs.com"}
	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)

	_, err := s.Create(
		ctx,
		&appb.CreateRequest{Name: "teresa"},
	)
	if err != nil {
		t.Error("Got error on Create: ", err)
	}
}

func TestCreateErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	s := NewService(fake)
	user := &storage.User{Email: "bad-user@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	_, err := s.Create(
		ctx,
		&appb.CreateRequest{Name: "teresa"},
	)
	if err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestCreateErrAppAlreadyExists(t *testing.T) {
	fake := NewFakeOperations()
	user := &storage.User{Email: "gopher@luizalabs.com"}
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)

	_, err := s.Create(
		ctx,
		&appb.CreateRequest{Name: name},
	)
	if err != ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %s", err)
	}
}
