package user

import (
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
)

func createFakeUser(db *gorm.DB, name, email, password string, isAdmin bool) error {
	p, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u := &database.User{
		Name:     name,
		Email:    email,
		Password: string(p),
		IsAdmin:  isAdmin}

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
	if err = createFakeUser(db, "Test", expectedEmail, expectedPassword, false); err != nil {
		t.Fatal("error on create fake user: ", err)
	}

	token, err := dbu.Login(expectedEmail, expectedPassword, time.Second)
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

	if _, err := dbu.Login("invalid@luizalabs.com", "secret", time.Second); err != auth.ErrPermissionDenied {
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
	if err = createFakeUser(db, "Test", expectedEmail, "", false); err != nil {
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

	users := map[string]database.User{
		"admin": database.User{Name: "Admin", Email: "sre@luizalabs.com", IsAdmin: true},
		"user1": database.User{Name: "User 1", Email: "teresa@luizalabs.com", IsAdmin: false},
		"user2": database.User{Name: "User 2", Email: "gopher@luizalabs.com", IsAdmin: false},
	}
	expectedPassword := "secret"

	for _, u := range users {
		if err = createFakeUser(db, u.Name, u.Email, "123456", u.IsAdmin); err != nil {
			t.Fatal("error on create fake user: ", err)
		}
	}

	testCases := []struct{ user, targetUser, userChanged string }{
		{user: "admin", targetUser: "", userChanged: "admin"},
		{user: "admin", targetUser: "user1", userChanged: "user1"},
		{user: "user2", targetUser: "user2", userChanged: "user2"},
	}

	for _, tc := range testCases {
		user := &database.User{Email: users[tc.user].Email, IsAdmin: users[tc.user].IsAdmin}
		if err = dbu.SetPassword(user, expectedPassword, users[tc.targetUser].Email); err != nil {
			t.Fatal("error trying to set a new password: ", err)
		}
		if _, err = dbu.Login(users[tc.userChanged].Email, expectedPassword, time.Second); err != nil {
			t.Error("error trying to make login with new password: ", err)
		}
	}
}

func TestDatabaseOperationsSetPasswordForInvalidTargetUser(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())

	testCases := []struct{ user, targetUser string }{
		{user: "gopher@luizalabs.com", targetUser: ""},
		{user: "gopher@luizalabs.com", targetUser: "teresa@luizalabs.com"},
		{user: "gopher@luizalabs.com", targetUser: "gopher@luizalabs.com"},
	}

	for _, tc := range testCases {
		user := &database.User{Email: tc.user, IsAdmin: true}
		if err := dbu.SetPassword(user, "123", tc.targetUser); err != ErrNotFound {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	}
}

func TestDatabaseOperationsSetPasswordErrPermissionDenied(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())

	users := map[string]database.User{
		"user1": database.User{Name: "User 1", Email: "teresa@luizalabs.com", IsAdmin: false},
		"user2": database.User{Name: "User 2", Email: "gopher@luizalabs.com", IsAdmin: false},
	}

	for _, u := range users {
		if err = createFakeUser(db, u.Name, u.Email, "123456", u.IsAdmin); err != nil {
			t.Fatal("error on create fake user: ", err)
		}
	}

	user := &database.User{Email: users["user1"].Email, IsAdmin: users["user1"].IsAdmin}
	if err = dbu.SetPassword(user, "123", users["user2"].Email); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestDatabaseOperationsDelete(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error opening in memory database ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())
	email := "teresa@luizalabs.com"
	if err := createFakeUser(db, "Test", email, "123456", false); err != nil {
		t.Fatal("error creating fake user: ", err)
	}
	if err := dbu.Delete(email); err != nil {
		t.Fatal("error deleting user: ", err)
	}
	if _, err := dbu.GetUser(email); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %s", err)
	}
}

func TestDatabaseOperationsDeleteUserNotFound(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error opening in memory database ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())
	if err := dbu.Delete("gopher@luizalabs.com"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %s", err)
	}
}

func TestDatabaseOperationsCreate(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error opening in memory database ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())
	name := "teresa"
	email := "teresa@luizalabs.com"
	pass := "test1234"
	if err := dbu.Create(name, email, pass, true); err != nil {
		t.Fatal("error creating user: ", err)
	}
	user, err := dbu.GetUser(email)
	if err != nil {
		t.Fatal("error fetching user: ", err)
	}
	if user.Name != name {
		t.Errorf("expected %s, got %s", name, user.Name)
	}
	if user.Email != email {
		t.Errorf("expected %s, got %s", email, user.Email)
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass))
	if err != nil {
		t.Fatal("error checking password: ", err)
	}
	if !user.IsAdmin {
		t.Fatal("expected true, got false")
	}
}

func TestDatabaseOperationsCreateUserAlreadyExists(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error opening in memory database ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())
	email := "teresa@luizalabs.com"

	if err := createFakeUser(db, "Test", email, "12345678", false); err != nil {
		t.Fatal("error creating fake user: ", err)
	}
	if err := dbu.Create("gopher", email, "12345678", false); err != ErrUserAlreadyExists {
		t.Errorf("expected ErrUserAlreadyExists, got %v", err)
	}
}

func TestDatabaseOperationsCreateUserInvalidPassword(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error opening in memory database ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())
	email := "teresa@luizalabs.com"

	if err := dbu.Create("gopher", email, "test", false); err != ErrInvalidPassword {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestDatabaseOperationsCreateUserInvalidEmail(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error opening in memory database ", err)
	}
	defer db.Close()

	dbu := NewDatabaseOperations(db, auth.NewFake())

	if err := dbu.Create("gopher", "gopher", "12345678", false); err != ErrInvalidEmail {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}
}
