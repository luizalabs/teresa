package team

import (
	"testing"

	context "golang.org/x/net/context"

	teampb "github.com/luizalabs/teresa/pkg/protobuf/team"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/user"
)

func TestTeamCreateSuccess(t *testing.T) {
	fake := NewFakeOperations()

	expectedEmail := "teresa@luizalabs.com"
	expectedName := "teresa"
	expectedURL := "http://teresa.io"

	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{Email: "gopher", IsAdmin: true})

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
	ctx := context.WithValue(context.Background(), "user", &database.User{IsAdmin: false})
	if _, err := s.Create(ctx, &teampb.CreateRequest{}); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestTeamCreateTeamAlreadyExists(t *testing.T) {
	fake := NewFakeOperations()

	expectedName := "teresa"
	fake.(*FakeOperations).Storage[expectedName] = &database.Team{Name: expectedName}

	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{Email: "gopher", IsAdmin: true})

	if _, err := s.Create(ctx, &teampb.CreateRequest{Name: expectedName}); err != ErrTeamAlreadyExists {
		t.Errorf("expected ErrTeamAlreadyExists, got %v", err)
	}
}

func TestTeamAddUserSuccess(t *testing.T) {
	fake := NewFakeOperations()

	expectedName := "teresa"
	expectedUserEmail := "gopher@luizalabs.com"
	fake.(*FakeOperations).Storage[expectedName] = &database.Team{Name: expectedName}
	fake.(*FakeOperations).UserOps.(*user.FakeOperations).Storage[expectedUserEmail] = &database.User{Email: expectedUserEmail}

	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{Email: "gopher@luizalabs.com", IsAdmin: true})

	req := &teampb.AddUserRequest{Name: expectedName, User: expectedUserEmail}
	if _, err := s.AddUser(ctx, req); err != nil {
		t.Fatal("Got error on make AddUser: ", err)
	}

	teamWithUser := fake.(*FakeOperations).Storage[expectedName]
	if len(teamWithUser.Users) == 0 {
		t.Fatal("AddUser dont add user for a team")
	}

	for _, u := range teamWithUser.Users {
		if u.Email == expectedUserEmail {
			return
		}
	}
	t.Errorf("AddUser didn't add user in team")
}

func TestTeamAddUserNotFound(t *testing.T) {
	fake := NewFakeOperations()
	s := NewService(fake)

	ctx := context.WithValue(context.Background(), "user", &database.User{Email: "gopher", IsAdmin: true})
	req := &teampb.AddUserRequest{Name: "teresa", User: "gopher"}
	if _, err := s.AddUser(ctx, req); err != ErrNotFound {
		t.Errorf("expected error ErrNotFound, got %v", err)
	}
}

func TestTeamAddUserUserNotFound(t *testing.T) {
	fake := NewFakeOperations()

	expectedName := "teresa"
	fake.(*FakeOperations).Storage[expectedName] = &database.Team{Name: expectedName}
	s := NewService(fake)

	ctx := context.WithValue(context.Background(), "user", &database.User{Email: "gopher", IsAdmin: true})
	req := &teampb.AddUserRequest{Name: "teresa", User: "gopher"}
	if _, err := s.AddUser(ctx, req); err != user.ErrNotFound {
		t.Errorf("expected error ErrNotFound, got %v", err)
	}
}

func TestTeamAddUserUserAlreadyInTeam(t *testing.T) {
	fake := NewFakeOperations()

	expectedName := "teresa"
	expectedUserEmail := "gopher@luizalabs.com"
	fake.(*FakeOperations).UserOps.(*user.FakeOperations).Storage[expectedUserEmail] = &database.User{Email: expectedUserEmail}
	fake.(*FakeOperations).Storage[expectedName] = &database.Team{
		Name:  expectedName,
		Users: []database.User{{Email: expectedUserEmail}},
	}

	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{Email: "gopher@luizalabs.com", IsAdmin: true})
	req := &teampb.AddUserRequest{Name: expectedName, User: expectedUserEmail}

	if _, err := s.AddUser(ctx, req); err != ErrUserAlreadyInTeam {
		t.Errorf("expected error ErrUserAlreadyInTeam, got %v", err)
	}
}

func TestTeamAddUserErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()

	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{IsAdmin: false})
	req := &teampb.AddUserRequest{Name: "teresa", User: "gopher"}

	if _, err := s.AddUser(ctx, req); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestTeamListUserAdmin(t *testing.T) {
	var testData = []struct {
		teamName   string
		usersEmail []string
	}{
		{teamName: "Empty"},
		{teamName: "teresa", usersEmail: []string{"gopher@luizalabs.com", "k8s@luizalabs.com"}},
	}

	fake := NewFakeOperations()
	for _, tc := range testData {
		fakeTeam := &database.Team{Name: tc.teamName}
		for _, email := range tc.usersEmail {
			fakeTeam.Users = append(fakeTeam.Users, database.User{Email: email})
		}
		fake.(*FakeOperations).Storage[tc.teamName] = fakeTeam
	}

	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{IsAdmin: true})
	resp, err := s.List(ctx, &teampb.Empty{})
	if err != nil {
		t.Fatal("error on list teams:", err)
	}

	if len(resp.Teams) != len(testData) {
		t.Errorf("expected %d, got %d", len(testData), len(resp.Teams))
	}

	var (
		emptyTeam  *teampb.ListResponse_Team
		teresaTeam *teampb.ListResponse_Team
	)

	if resp.Teams[0].Name == "Empty" {
		emptyTeam = resp.Teams[0]
		teresaTeam = resp.Teams[1]
	} else {
		teresaTeam = resp.Teams[0]
		emptyTeam = resp.Teams[1]
	}

	if len(emptyTeam.Users) != 0 {
		t.Errorf("expected 0, got %d", len(emptyTeam.Users))
	}
	if len(teresaTeam.Users) != 2 {
		t.Errorf("expected 2, got %d", len(teresaTeam.Users))
	}
}

func TestTeamList(t *testing.T) {
	expectedUserEmail := "gopher@luizalabs.com"
	var testData = []struct {
		teamName   string
		usersEmail []string
	}{
		{teamName: "Empty"},
		{teamName: "vimmers", usersEmail: []string{"k8s@luizalabs.com"}},
		{teamName: "teresa", usersEmail: []string{expectedUserEmail, "k8s@luizalabs.com"}},
		{teamName: "gophers", usersEmail: []string{expectedUserEmail, "pike@luizalabs.com", "cheney@luizalabs.com"}},
	}

	fake := NewFakeOperations()
	for _, tc := range testData {
		fakeTeam := &database.Team{Name: tc.teamName}
		for _, email := range tc.usersEmail {
			fakeTeam.Users = append(fakeTeam.Users, database.User{Email: email})
		}
		fake.(*FakeOperations).Storage[tc.teamName] = fakeTeam
	}

	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{IsAdmin: false, Email: expectedUserEmail})
	resp, err := s.List(ctx, &teampb.Empty{})
	if err != nil {
		t.Fatal("error on list teams:", err)
	}
	if len(resp.Teams) != 2 {
		t.Errorf("expected 2, got %d", len(resp.Teams))
	}
}

func TestRemoveUserSuccess(t *testing.T) {
	fake := NewFakeOperations()
	expectedTeam := "teresa"
	expectedUserEmail := "gopher@luizalabs.com"
	fake.(*FakeOperations).UserOps.(*user.FakeOperations).Storage[expectedUserEmail] = &database.User{Email: expectedUserEmail}
	fake.(*FakeOperations).Storage[expectedTeam] = &database.Team{
		Name:  expectedTeam,
		Users: []database.User{{Email: expectedUserEmail}},
	}
	srv := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{Email: "gopher@luizalabs.com", IsAdmin: true})
	req := &teampb.RemoveUserRequest{Team: expectedTeam, User: expectedUserEmail}

	if _, err := srv.RemoveUser(ctx, req); err != nil {
		t.Fatal("got error on RemoveUser: ", err)
	}

	teamWithoutUser := fake.(*FakeOperations).Storage[expectedTeam]

	if len(teamWithoutUser.Users) != 0 {
		t.Errorf("RemoveUser didn't remove user from team")
	}
}

func TestRemoveUserTeamNotFound(t *testing.T) {
	fake := NewFakeOperations()
	srv := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{Email: "gopher", IsAdmin: true})
	req := &teampb.RemoveUserRequest{Team: "teresa", User: "gopher"}

	if _, err := srv.RemoveUser(ctx, req); err != ErrNotFound {
		t.Errorf("expected error ErrNotFound, got %v", err)
	}
}

func TestRemoveUserNotFound(t *testing.T) {
	fake := NewFakeOperations()
	expectedTeam := "teresa"
	fake.(*FakeOperations).Storage[expectedTeam] = &database.Team{Name: expectedTeam}
	srv := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{Email: "gopher", IsAdmin: true})
	req := &teampb.RemoveUserRequest{Team: "teresa", User: "gopher"}

	if _, err := srv.RemoveUser(ctx, req); err != user.ErrNotFound {
		t.Errorf("expected error user.ErrNotFound, got %v", err)
	}
}

func TestRemoveUserNotInTeam(t *testing.T) {
	fake := NewFakeOperations()
	expectedTeam := "teresa"
	expectedUserEmail := "gopher@luizalabs.com"
	fake.(*FakeOperations).UserOps.(*user.FakeOperations).Storage[expectedUserEmail] = &database.User{Email: expectedUserEmail}
	fake.(*FakeOperations).Storage[expectedTeam] = &database.Team{Name: expectedTeam}
	srv := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{Email: "gopher@luizalabs.com", IsAdmin: true})
	req := &teampb.RemoveUserRequest{Team: expectedTeam, User: expectedUserEmail}

	if _, err := srv.RemoveUser(ctx, req); err != ErrUserNotInTeam {
		t.Errorf("expected error ErrUserNotInTeam, got %v", err)
	}
}

func TestRemoveUserPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	srv := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{IsAdmin: false})
	req := &teampb.RemoveUserRequest{Team: "teresa", User: "gopher"}

	if _, err := srv.RemoveUser(ctx, req); err != auth.ErrPermissionDenied {
		t.Errorf("expected auth.ErrPermissionDenied, got %v", err)
	}
}

func TestTeamRenameSuccess(t *testing.T) {
	fake := NewFakeOperations()

	expectedEmail := "teresa@luizalabs.com"
	oldName := "teresa"
	fake.(*FakeOperations).Storage[oldName] = &database.Team{Name: oldName, Email: expectedEmail}

	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{Email: "gopher", IsAdmin: true})

	newName := "gophers"
	req := &teampb.RenameRequest{OldName: oldName, NewName: newName}
	if _, err := s.Rename(ctx, req); err != nil {
		t.Fatal("Got error renaming team:", err)
	}

	oldTeam := fake.(*FakeOperations).Storage[newName]
	if oldTeam.Email != expectedEmail {
		t.Errorf("expected %s, got %s", expectedEmail, oldTeam.Email)
	}
}

func TestTeamRenameErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()

	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{IsAdmin: false})
	if _, err := s.Rename(
		ctx, &teampb.RenameRequest{OldName: "old", NewName: "new"},
	); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestTeamRenameTeamAlreadyExists(t *testing.T) {
	fake := NewFakeOperations()

	oldName := "gopher"
	createdName := "teresa"
	fake.(*FakeOperations).Storage[oldName] = &database.Team{Name: oldName}
	fake.(*FakeOperations).Storage[createdName] = &database.Team{Name: createdName}

	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", &database.User{Email: "gopher", IsAdmin: true})

	req := &teampb.RenameRequest{OldName: oldName, NewName: createdName}
	if _, err := s.Rename(ctx, req); err != ErrTeamAlreadyExists {
		t.Errorf("expected ErrTeamAlreadyExists, got %v", err)
	}
}
