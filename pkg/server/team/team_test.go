package team

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/user"
)

func createFakeTeam(db *gorm.DB, name, email, url string) error {
	t := &storage.Team{
		Name:  name,
		Email: email,
		URL:   url}
	return db.Create(t).Error
}

func TestDatabaseOperationsCreate(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	dbt := NewDatabaseOperations(db, user.NewFakeOperations())

	expectedEmail := "teresa@luizalabs.com"
	expectedName := "teresa"
	expectedURL := "http://teresa.io"

	if err = dbt.Create(expectedName, expectedEmail, expectedURL); err != nil {
		t.Fatal("error trying to create a team", err)
	}

	newTeam, err := dbt.(*DatabaseOperations).getTeam(expectedName)
	if err != nil {
		t.Fatal("error on get team:", err)
	}

	if newTeam.Name != expectedName {
		t.Errorf("expected %s, got %s", expectedName, newTeam.Name)
	}
	if newTeam.Email != expectedEmail {
		t.Errorf("expected %s, got %s", expectedEmail, newTeam.Email)
	}
	if newTeam.URL != expectedURL {
		t.Errorf("expected %s, got %s", expectedURL, newTeam.URL)
	}
}

func TestDatabaseOperationsCreateTeamAlreadyExists(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	dbt := NewDatabaseOperations(db, user.NewFakeOperations())

	teamName := "teresa"
	if err := createFakeTeam(db, teamName, "", ""); err != nil {
		t.Fatal("error on create a fake team:", err)
	}

	if err = dbt.Create(teamName, "", ""); err != ErrTeamAlreadyExists {
		t.Errorf("expected ErrTeamAlreadyExists, got %v", err)
	}
}

func TestDatabaseOperationsAddUser(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	db.AutoMigrate(&storage.User{})
	defer db.Close()

	expectedUserEmail := "gopher"

	dbt := NewDatabaseOperations(db, user.NewFakeOperations())
	dbt.(*DatabaseOperations).UserOps.(*user.FakeOperations).Storage[expectedUserEmail] = &storage.User{Email: expectedUserEmail}

	expectedTeam := "teresa"
	if err := dbt.Create(expectedTeam, "", ""); err != nil {
		t.Fatal("error on create a team:", err)
	}

	if err := dbt.AddUser(expectedTeam, expectedUserEmail); err != nil {
		t.Errorf("error trying on add user to a team: %v", err)
	}
}

func TestDatabaseOperationsAddUserTeamNotFound(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	dbt := NewDatabaseOperations(db, user.NewFakeOperations())
	if err := dbt.AddUser("teresa", "gopher"); err != ErrNotFound {
		t.Errorf("expected error ErrNotFound, got %v", err)
	}
}

func TestDatabaseOperationsAddUserUserNotFound(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	db.AutoMigrate(&storage.User{})
	defer db.Close()

	dbt := NewDatabaseOperations(db, user.NewFakeOperations())

	expectedTeam := "teresa"
	if err := dbt.Create(expectedTeam, "", ""); err != nil {
		t.Fatal("error on create a team:", err)
	}

	if err := dbt.AddUser(expectedTeam, "gopher"); err != user.ErrNotFound {
		t.Errorf("expected error ErrNotFound, got %v", err)
	}
}

func TestDatabaseOperationsAddUserUserAlreadyInTeam(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	db.AutoMigrate(&storage.User{})
	defer db.Close()

	expectedUserEmail := "gopher"

	dbt := NewDatabaseOperations(db, user.NewFakeOperations())
	dbt.(*DatabaseOperations).UserOps.(*user.FakeOperations).Storage[expectedUserEmail] = &storage.User{Email: expectedUserEmail}

	expectedTeam := "teresa"
	if err := dbt.Create(expectedTeam, "", ""); err != nil {
		t.Fatal("error on create a team:", err)
	}

	for _, expectedErr := range []error{nil, ErrUserAlreadyInTeam} {
		if err := dbt.AddUser(expectedTeam, expectedUserEmail); err != expectedErr {
			t.Errorf("expected %v, got %v", expectedErr, err)
		}
	}
}
