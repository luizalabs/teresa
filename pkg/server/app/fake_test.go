package app

import (
	"testing"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

func TestFakeOperationsCreate(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	user := &storage.User{Name: "gopher@luizalabs.com"}
	app := &App{Name: name}

	err := fake.Create(user, app)
	if err != nil {
		t.Fatal("Error creating app in FakeOperations: ", err)
	}

	fakeApp := fake.(*FakeOperations).Storage[name]
	if fakeApp == nil {
		t.Fatal("expected a valid app, got nil")
	}
	if fakeApp.Name != name {
		t.Errorf("expected %s, got %s", name, fakeApp.Name)
	}
}

func TestFakeOperationsCreateErrAppAlreadyExists(t *testing.T) {
	fake := NewFakeOperations()
	user := &storage.User{Email: "gopher@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage["teresa"] = app

	if err := fake.Create(user, app); err != ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestFakeOperationsCreateErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	user := &storage.User{Email: "bad-user@luizalabs.com"}
	app := &App{Name: "teresa"}

	if err := fake.Create(user, app); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}
