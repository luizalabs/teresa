package team

import (
	"errors"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa-api/models/storage"
)

func getTeam(db *gorm.DB, name string) (*storage.Team, error) {
	t := new(storage.Team)
	if db.Where(&storage.Team{Name: name}).First(t).RecordNotFound() {
		return nil, errors.New("Team not found")
	}
	return t, nil
}

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

	dbt := NewDatabaseOperations(db)

	expectedEmail := "teresa@luizalabs.com"
	expectedName := "teresa"
	expectedURL := "http://teresa.io"

	if err = dbt.Create(expectedName, expectedEmail, expectedURL); err != nil {
		t.Fatal("error trying to create a team", err)
	}

	newTeam, err := getTeam(db, expectedName)
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

	dbt := NewDatabaseOperations(db)

	teamName := "teresa"
	if err := createFakeTeam(db, teamName, "", ""); err != nil {
		t.Fatal("error on create a fake team:", err)
	}

	if err = dbt.Create(teamName, "", ""); err != ErrTeamAlreadyExists {
		t.Errorf("expected ErrTeamAlreadyExists, got %v", err)
	}
}
