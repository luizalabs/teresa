package user

import (
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

func createFakeUser(db *gorm.DB, email, password string) error {
	p, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u := &storage.User{
		Name:     "Test",
		Email:    email,
		Password: string(p),
		IsAdmin:  false}

	return db.Create(u).Error
}

func TestDatabaseOperationsLogin(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())

	expectedEmail := "teresa@luizalabs.com"
	expectedPassword := "secret"
	if err = createFakeUser(db, expectedEmail, expectedPassword); err != nil {
		t.Fatal("error on create fake user: ", err)
	}

	token, err := dbu.Login(expectedEmail, expectedPassword)
	if err != nil {
		t.Fatal("Error on perform Login: ", err)
	}
	if token == "" {
		t.Error("expected a valid token, got a blank string")
	}
}

func TestDatabaseOperationsBadLogin(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())

	if _, err := dbu.Login("invalid@luizalabs.com", "secret"); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestDatabaseOperationsGetUser(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())

	expectedEmail := "teresa@luizalabs.com"
	if err = createFakeUser(db, expectedEmail, ""); err != nil {
		t.Fatal("error on create fake user: ", err)
	}

	u, err := dbu.GetUser(expectedEmail)
	if err != nil {
		t.Fatal("error on GetUser: ", err)
	}
	if u.Email != expectedEmail {
		t.Errorf("expected get user with e-mail %s, got %s", expectedEmail, u.Email)
	}
}

func TestDatabaseOperationsGetUserNotFound(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())
	if _, err := dbu.GetUser("gopher@luizalabs.com"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %s", err)
	}
}

func TestDatabaseOperationsSetPassword(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())

	expectedEmail := "teresa@luizalabs.com"
	expectedPassword := "secret"
	if err = createFakeUser(db, expectedEmail, "123456"); err != nil {
		t.Fatal("error on create fake user: ", err)
	}

	if err = dbu.SetPassword(expectedEmail, expectedPassword); err != nil {
		t.Fatal("error trying to set a new password: ", err)
	}
	if _, err = dbu.Login(expectedEmail, expectedPassword); err != nil {
		t.Error("error trying to make login with new password: ", err)
	}
}

func TestDatabaseOperationsSetPasswordForInvalidUser(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())
	if err := dbu.SetPassword("gopher@luizalabs.com", "123"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
