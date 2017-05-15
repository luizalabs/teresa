package team

import (
	"testing"

	context "golang.org/x/net/context"

	teampb "github.com/luizalabs/teresa-api/pkg/protobuf/team"
	"github.com/luizalabs/teresa-api/pkg/server/auth"

	"github.com/luizalabs/teresa-api/models/storage"
)

func TestTeamCreateSuccess(t *testing.T) {
	fake := NewFakeOperations()

	expectedEmail := "teresa@luizalabs.com"
	expectedName := "teresa"
	expectedURL := "http://teresa.io"

	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &storage.User{Email: "gopher", IsAdmin: true})

	req := &teampb.CreateRequest{Name: expectedName, Email: expectedEmail, Url: expectedURL}
	if _, err := s.Create(ctx, req); err != nil {
		t.Fatal("Got error on make Create: ", err)
	}

	newTeam := fake.(*FakeOperations).Storage[expectedName]
	if newTeam.Email != expectedEmail {
		t.Errorf("expected %s, got %s", expectedEmail, newTeam.Email)
	}
}

func TestTeamCreateErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()

	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &storage.User{IsAdmin: false})
	if _, err := s.Create(ctx, &teampb.CreateRequest{}); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestTeamCreateTeamAlreadyExists(t *testing.T) {
	fake := NewFakeOperations()

	expectedName := "teresa"
	fake.(*FakeOperations).Storage[expectedName] = &storage.Team{Name: expectedName}

	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &storage.User{Email: "gopher", IsAdmin: true})

	if _, err := s.Create(ctx, &teampb.CreateRequest{Name: expectedName}); err != ErrTeamAlreadyExists {
		t.Errorf("expected ErrTeamAlreadyExists, got %v", err)
	}
}
