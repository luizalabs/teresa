package team

import (
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
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

func TestDatabaseOperationsListWithoutTeams(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	dbt := NewDatabaseOperations(db, user.NewFakeOperations())
	teams, err := dbt.List()
	if err != nil {
		t.Error("error on list teams:", err)
	}
	if len(teams) > 0 {
		t.Errorf("expected 0, got %d", len(teams))
	}
}

func TestDatabaseOperationsList(t *testing.T) {
	var testData = []struct {
		teamName   string
		usersEmail []string
	}{
		{teamName: "Empty"},
		{teamName: "teresa", usersEmail: []string{"gopher@luizalabs.com", "k8s@luizalabs.com"}},
	}

	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	uOps := user.NewDatabaseOperations(db, auth.NewFake())
	dbt := NewDatabaseOperations(db, uOps)
	for _, tc := range testData {
		if err := dbt.Create(tc.teamName, "", ""); err != nil {
			t.Fatal("error on create team:", err)
		}
		for _, email := range tc.usersEmail {
			if err := uOps.Create(email, email, "12345678", false); err != nil {
				t.Fatal("error on create user", err)
			}
			if err := dbt.AddUser(tc.teamName, email); err != nil {
				t.Fatal("error on add user to team: ", err)
			}
		}
	}

	teams, err := dbt.List()
	if err != nil {
		t.Fatal("error on list teams:", err)
	}
	if len(teams) != len(testData) {
		t.Fatalf("expected %d, got %d", len(testData), len(teams))
	}

	for i := range teams {
		if teams[i].Name != testData[i].teamName {
			t.Errorf("expected %s, got %s", testData[i].teamName, teams[i].Name)
		}
		if len(teams[i].Users) != len(testData[i].usersEmail) {
			t.Fatalf(
				"expected %d users in team, got %d",
				len(testData[i].usersEmail),
				len(teams[i].Users),
			)
		}
		for idx := range teams[i].Users {
			if teams[i].Users[idx].Email != testData[i].usersEmail[idx] {
				t.Errorf(
					"expected %s, got %s",
					testData[i].usersEmail[idx],
					teams[i].Users[idx].Email,
				)
			}
		}
	}
}

func TestDatabaseOperationsListByUser(t *testing.T) {
	expectedUserEmail := "gopher@luizalabs.com"

	var testData = []struct {
		teamName   string
		usersEmail []string
	}{
		{teamName: "Empty"},
		{teamName: "teresa", usersEmail: []string{expectedUserEmail, "k8s@luizalabs.com"}},
		{teamName: "gophers", usersEmail: []string{expectedUserEmail, "john@luizalabs.com"}},
		{teamName: "vimers", usersEmail: []string{"k8s@luizalabs.com", "john@luizalabs.com"}},
	}

	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	uOps := user.NewDatabaseOperations(db, auth.NewFake())
	dbt := NewDatabaseOperations(db, uOps)
	for _, tc := range testData {
		if err := dbt.Create(tc.teamName, "", ""); err != nil {
			t.Fatal("error on create team:", err)
		}
		for _, email := range tc.usersEmail {
			err := uOps.Create(email, email, "12345678", false)
			if err != nil && err != user.ErrUserAlreadyExists {
				t.Fatal("error on create user", err)
			}
			if err := dbt.AddUser(tc.teamName, email); err != nil {
				t.Fatal("error on add user to team: ", err)
			}
		}
	}

	teams, err := dbt.ListByUser(expectedUserEmail)
	if err != nil {
		t.Fatal("error on list teams:", err)
	}
	if len(teams) != 2 {
		t.Fatalf("expected 2, got %d", len(teams))
	}

	for _, currentTeam := range teams {
		if currentTeam.Name != "gophers" && currentTeam.Name != "teresa" {
			t.Errorf("expected gophers or teresa, got %s", currentTeam.Name)
		}
		if len(currentTeam.Users) != 2 {
			t.Errorf("expected 2, got %d", len(currentTeam.Users))
		}
	}
}

func TestDatabaseOperationsListByUserWithoutTeams(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	defer db.Close()

	expectedUserEmail := "gopher@luizalabs.com"

	uOps := user.NewDatabaseOperations(db, auth.NewFake())
	dbt := NewDatabaseOperations(db, uOps)

	if err := uOps.Create("", expectedUserEmail, "12345678", false); err != nil {
		t.Fatal("error on creating user:", err)
	}

	teams, err := dbt.ListByUser(expectedUserEmail)
	if err != nil {
		t.Error("error on list teams:", err)
	}
	if len(teams) > 0 {
		t.Errorf("expected 0, got %d", len(teams))
	}
}
