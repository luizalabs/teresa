package app

import (
	"testing"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
	"github.com/luizalabs/teresa-api/pkg/server/team"
)

type fakeK8sOperations struct{}

type errK8sOperations struct{ Err error }

func (*fakeK8sOperations) Create(app *App, st st.Storage) error {
	return nil
}

func (e *errK8sOperations) Create(app *App, st st.Storage) error {
	return e.Err
}

func TestAppOperationsCreate(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewAppOperations(tops, &fakeK8sOperations{}, nil)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
	}

	if err := ops.Create(user, app); err != nil {
		t.Fatal("error creating app: ", err)
	}
}

func TestAppOperationsCreateErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewAppOperations(tops, &fakeK8sOperations{}, nil)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}

	if err := ops.Create(user, app); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestAppOperationsCreateErrAppAlreadyExists(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewAppOperations(tops, &fakeK8sOperations{}, nil)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
	}

	if err := ops.Create(user, app); err != nil {
		t.Fatal("error creating app: ", err)
	}

	ops.(*AppOperations).kops = &errK8sOperations{Err: ErrAlreadyExists}
	if err := ops.Create(user, app); err != ErrAlreadyExists {
		t.Errorf("expected ErrAppAlreadyExists got %s", err)
	}
}
