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
		t.Fatal("error on open in memory databaser ", err)
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
		t.Fatal("error on open in memory databaser ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())

	if _, err := dbu.Login("invalid@luizalabs.com", "secret"); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}
