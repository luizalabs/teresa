package user

import (
	"testing"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

func TestFakeOperationsLogin(t *testing.T) {
	fake := NewFakeOperations()

	expectedEmail := "teresa@luizalabs.com"
	expectedPassword := "123456"
	fake.(*FakeOperations).Storage[expectedEmail] = &storage.User{
		Password: expectedPassword,
		Email:    expectedEmail,
	}

	token, err := fake.Login(expectedEmail, expectedPassword)
	if err != nil {
		t.Fatal("Error on perform Login in FakeOperations: ", err)
	}
	expectedToken := "good token"
	if token != expectedToken {
		t.Errorf("expected %s, got %s", expectedToken, token)
	}
}

func TestFakeOperationsBadLogin(t *testing.T) {
	fake := NewFakeOperations()

	if _, err := fake.Login("invalid@luizalabs.com", "foo"); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestFakeOperationsGetUser(t *testing.T) {
	fake := NewFakeOperations()

	expectedEmail := "teresa@luizalabs.com"
	fake.(*FakeOperations).Storage[expectedEmail] = &storage.User{
		Password: "foo",
		Email:    expectedEmail,
	}

	u, err := fake.GetUser(expectedEmail)
	if err != nil {
		t.Fatal("error on get user: ", err)
	}
	if u.Email != expectedEmail {
		t.Errorf("expected %s, got %s", expectedEmail, u.Email)
	}
}

func TestFakeOperationsGetUserNotFound(t *testing.T) {
	fake := NewFakeOperations()
	if _, err := fake.GetUser("gopher@luizalabs.com"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %s", err)
	}
}

func TestFakeOperationsSetPassword(t *testing.T) {
	fake := NewFakeOperations()

	expectedEmail := "teresa@luizalabs.com"
	expectedPassword := "123456"
	fake.(*FakeOperations).Storage[expectedEmail] = &storage.User{
		Password: "gopher",
		Email:    expectedEmail,
	}

	err := fake.SetPassword(expectedEmail, expectedPassword)
	if err != nil {
		t.Fatal("error trying to change user password: ", err)
	}
	currentPassword := fake.(*FakeOperations).Storage[expectedEmail].Password
	if currentPassword != expectedPassword {
		t.Errorf("expected %s, got %s", expectedPassword, currentPassword)
	}
}

func TestFakeOperationsSetPasswordUserNotFound(t *testing.T) {
	fake := NewFakeOperations()

	if err := fake.SetPassword("gopher@luizalabs.com", "123"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFakeOperationsDelete(t *testing.T) {
	fake := NewFakeOperations()

	email := "teresa@luizalabs.com"
	fake.(*FakeOperations).Storage[email] = &storage.User{
		Password: "gopher",
		Email:    email,
	}

	if err := fake.Delete(email); err != nil {
		t.Fatal("Error performing delete in FakeOperations: ", err)
	}
	_, ok := fake.(*FakeOperations).Storage[email]
	if ok {
		t.Errorf("expected false for key %s, got true", email)
	}
}

func TestFakeOperationsDeleteUserNotFound(t *testing.T) {
	fake := NewFakeOperations()

	if err := fake.Delete("gopher@luizalabs.com"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
