package app

import (
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
)

type fakeK8sOperations struct{}

type errK8sOperations struct{ Err error }

func (*fakeK8sOperations) Create(app *App, st st.Storage) error {
	return nil
}

func (e *errK8sOperations) Create(app *App, st st.Storage) error {
	return e.Err
}

func createFakeUser(db *gorm.DB, email string) (*storage.User, error) {
	user := &storage.User{Email: email}
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func createFakeTeam(db *gorm.DB, name string) (*storage.Team, error) {
	team := &storage.Team{Name: name}
	if err := db.Create(team).Error; err != nil {
		return nil, err
	}
	return team, nil
}

func addUserTeam(db *gorm.DB, user *storage.User, team *storage.Team) error {
	return db.Model(team).Association("Users").Append(user).Error
}

func TestAppOperationsCreate(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	db.AutoMigrate(&storage.User{}, &storage.Team{})
	defer db.Close()
	ops := NewAppOperations(db, &fakeK8sOperations{}, nil)
	name := "luizalabs"
	app := &App{Name: "teresa", Team: name}

	user, err := createFakeUser(db, "teresa@luizalabs.com")
	if err != nil {
		t.Fatal("error creating fake user: ", err)
	}
	team, err := createFakeTeam(db, name)
	if err != nil {
		t.Fatal("error creating fake team: ", err)
	}
	if err := addUserTeam(db, user, team); err != nil {
		t.Fatal("error adding fake user to team: ", err)
	}
	if err := ops.Create(user, app); err != nil {
		t.Fatal("error creating app: ", err)
	}
}

func TestAppOperationsCreateErrPermissionDenied(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	db.AutoMigrate(&storage.User{}, &storage.Team{})
	defer db.Close()
	ops := NewAppOperations(db, &fakeK8sOperations{}, nil)
	name := "luizalabs"
	app := &App{Name: "teresa", Team: name}

	user, err := createFakeUser(db, "teresa@luizalabs.com")
	if err != nil {
		t.Fatal("error creating fake user: ", err)
	}
	if _, err := createFakeTeam(db, name); err != nil {
		t.Fatal("error creating fake team: ", err)
	}
	if err := ops.Create(user, app); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestAppOperationsCreateErrAppAlreadyExists(t *testing.T) {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal("error on open in memory database ", err)
	}
	db.AutoMigrate(&storage.User{}, &storage.Team{})
	defer db.Close()
	ops := NewAppOperations(db, &fakeK8sOperations{}, nil)
	name := "luizalabs"
	app := &App{Name: "teresa", Team: name}

	user, err := createFakeUser(db, "teresa@luizalabs.com")
	if err != nil {
		t.Fatal("error creating fake user: ", err)
	}
	team, err := createFakeTeam(db, name)
	if err != nil {
		t.Fatal("error creating fake team: ", err)
	}
	if err := addUserTeam(db, user, team); err != nil {
		t.Fatal("error adding fake user to team: ", err)
	}
	if err := ops.Create(user, app); err != nil {
		t.Fatal("error creating app: ", err)
	}

	ops.(*AppOperations).kops = &errK8sOperations{Err: ErrAlreadyExists}
	if err := ops.Create(user, app); err != ErrAlreadyExists {
		t.Errorf("expected ErrAppAlreadyExists got %s", err)
	}
}
