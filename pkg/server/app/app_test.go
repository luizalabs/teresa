package app

import (
	"errors"
	"testing"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
	"github.com/luizalabs/teresa-api/pkg/server/team"
)

type fakeK8sOperations struct{}

type errK8sOperations struct {
	NamespaceErr error
	QuotaErr     error
	SecretErr    error
	AutoScaleErr error
}

func (*fakeK8sOperations) CreateNamespace(app *App, user string) error {
	return nil
}

func (*fakeK8sOperations) CreateQuota(app *App) error {
	return nil
}

func (*fakeK8sOperations) CreateSecret(appName, secretName string, data map[string][]byte) error {
	return nil
}

func (*fakeK8sOperations) CreateAutoScale(app *App) error {
	return nil
}

func (e *errK8sOperations) CreateNamespace(app *App, user string) error {
	return e.NamespaceErr
}

func (e *errK8sOperations) CreateQuota(app *App) error {
	return e.QuotaErr
}

func (e *errK8sOperations) CreateSecret(appName, secretName string, data map[string][]byte) error {
	return e.SecretErr
}

func (e *errK8sOperations) CreateAutoScale(app *App) error {
	return e.AutoScaleErr
}

func TestAppOperationsCreate(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewAppOperations(tops, &fakeK8sOperations{}, fakeSt)
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
	fakeSt := st.NewFake()
	ops := NewAppOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}

	if err := ops.Create(user, app); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestAppOperationsCreateErrAppAlreadyExists(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewAppOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{NamespaceErr: ErrAlreadyExists}

	if err := ops.Create(user, app); err != ErrAlreadyExists {
		t.Errorf("expected ErrAppAlreadyExists got %s", err)
	}
}

func TestAppOperationsCreateErrQuota(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewAppOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{QuotaErr: errors.New("Quota Error")}

	if err := ops.Create(user, app); err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestAppOperationsCreateErrSecret(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewAppOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{SecretErr: errors.New("Secret Error")}

	if err := ops.Create(user, app); err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestAppOperationsCreateErrAutoScale(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewAppOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{AutoScaleErr: errors.New("AutoScale Error")}

	if err := ops.Create(user, app); err == nil {
		t.Errorf("expected error, got nil")
	}
}
