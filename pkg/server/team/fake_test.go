package team

import (
	"testing"

	"github.com/luizalabs/teresa-api/models/storage"
)

func TestFakeOperationsCreate(t *testing.T) {
	fake := NewFakeOperations()

	expectedEmail := "teresa@luizalabs.com"
	expectedName := "teresa"
	expectedURL := "http://teresa.io"

	if err := fake.Create(expectedName, expectedEmail, expectedURL); err != nil {
		t.Fatal("error trying to create a fake team", err)
	}

	fakeTeam := fake.(*FakeOperations).Storage[expectedName]
	if fakeTeam == nil {
		t.Fatal("expected a valid team, got nil")
	}

	if fakeTeam.Name != expectedName {
		t.Errorf("expected %s, got %s", expectedName, fakeTeam.Name)
	}
	if fakeTeam.Email != expectedEmail {
		t.Errorf("expected %s, got %s", expectedEmail, fakeTeam.Email)
	}
	if fakeTeam.URL != expectedURL {
		t.Errorf("expected %s, got %s", expectedURL, fakeTeam.URL)
	}
}

func TestFakeOperationsCreateTeamAlreadyExists(t *testing.T) {
	fake := NewFakeOperations()

	teamName := "teresa"
	fake.(*FakeOperations).Storage[teamName] = &storage.Team{Name: teamName}

	if err := fake.Create(teamName, "", ""); err != ErrTeamAlreadyExists {
		t.Errorf("expected ErrTeamAlreadyExists, got %v", err)
	}
}
