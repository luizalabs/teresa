package app

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"k8s.io/client-go/pkg/api"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
	"github.com/luizalabs/teresa-api/pkg/server/team"
)

type fakeK8sOperations struct{}

type errK8sOperations struct {
	Err          error
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

func (*fakeK8sOperations) PodList(namespace string) ([]*Pod, error) {
	pl := []*Pod{
		{Name: "pod 1", State: string(api.PodRunning)},
		{Name: "pod 2", State: string(api.PodRunning)},
	}
	return pl, nil
}

func (*fakeK8sOperations) PodLogs(namespace, podName string, lines int64, follow bool) (io.ReadCloser, error) {
	r := bytes.NewBufferString("foo\nbar")
	return ioutil.NopCloser(r), nil
}

func (*fakeK8sOperations) NamespaceAnnotation(namespace, annotation string) (string, error) {
	return "", nil
}

func (*fakeK8sOperations) NamespaceLabel(namespace, label string) (string, error) {
	return "luizalabs", nil
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

func (e *errK8sOperations) PodList(namespace string) ([]*Pod, error) {
	return nil, e.Err
}

func (e *errK8sOperations) PodLogs(namespace, podName string, lines int64, follow bool) (io.ReadCloser, error) {
	return nil, e.Err
}

func (e *errK8sOperations) NamespaceAnnotation(namespace, annotation string) (string, error) {
	return "", e.Err
}

func (e *errK8sOperations) NamespaceLabel(namespace, label string) (string, error) {
	return "", e.Err
}

func TestAppOperationsCreate(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
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
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
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
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
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

func TestAppOperationsLogs(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
	}

	rc, err := ops.Logs(user, app.Name, 10, false)
	if err != nil {
		t.Fatal("error on get logs: ", err)
	}
	defer rc.Close()

	count := 0
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		text := scanner.Text()
		if !strings.HasPrefix(text, "[pod ") {
			t.Errorf("expected log with pod name, got %s", text)
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal("error on read logs:", err)
	}

	if count != 4 { // see fakeK8sOperations.PodLogs
		t.Errorf("expected 4, got %d", count)
	}
}

func TestAppOperationsLogsErrPermissionDenied(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &fakeK8sOperations{}, nil)
	user := &storage.User{Email: "teresa@luizalabs.com"}

	if _, err := ops.Logs(user, "teresa", 10, false); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestAppOperationsLogsErrNotFound(t *testing.T) {
	tops := team.NewFakeOperations()
	ops := NewOperations(tops, &errK8sOperations{Err: ErrNotFound}, nil)
	user := &storage.User{Email: "teresa@luizalabs.com"}

	if _, err := ops.Logs(user, "teresa", 10, false); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAppOperationsCreateErrQuota(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{QuotaErr: errors.New("Quota Error")}

	if ops.Create(user, app) == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestAppOperationsCreateErrSecret(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{SecretErr: errors.New("Secret Error")}

	if ops.Create(user, app) == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestAppOperationsCreateErrAutoScale(t *testing.T) {
	tops := team.NewFakeOperations()
	fakeSt := st.NewFake()
	ops := NewOperations(tops, &fakeK8sOperations{}, fakeSt)
	name := "luizalabs"
	user := &storage.User{Email: "teresa@luizalabs.com"}
	app := &App{Name: "teresa", Team: name}
	tops.(*team.FakeOperations).Storage[name] = &storage.Team{
		Name:  name,
		Users: []storage.User{*user},
	}
	ops.(*AppOperations).kops = &errK8sOperations{AutoScaleErr: errors.New("AutoScale Error")}

	if ops.Create(user, app) == nil {
		t.Errorf("expected error, got nil")
	}
}
