package app

import (
	"bufio"
	"testing"

	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
)

func TestFakeOperationsCreate(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	user := &database.User{Name: "gopher@luizalabs.com"}
	app := &App{Name: name}

	err := fake.Create(user, app)
	if err != nil {
		t.Fatal("Error creating app in FakeOperations: ", err)
	}

	fakeApp := fake.(*FakeOperations).Storage[name]
	if fakeApp == nil {
		t.Fatal("expected a valid app, got nil")
	}
	if fakeApp.Name != name {
		t.Errorf("expected %s, got %s", name, fakeApp.Name)
	}
}

func TestFakeOperationsCreateErrAppAlreadyExists(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Email: "gopher@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage["teresa"] = app

	if err := fake.Create(user, app); err != ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestFakeOperationsCreateErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Email: "bad-user@luizalabs.com"}
	app := &App{Name: "teresa"}

	if err := fake.Create(user, app); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestFakeOperationsLogs(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Name: "gopher@luizalabs.com"}
	app := &App{Name: "teresa"}

	fake.(*FakeOperations).Storage[app.Name] = app

	expectedLines := 10
	rc, err := fake.Logs(user, app.Name, int64(expectedLines), false)
	if err != nil {
		t.Fatal("error on get logs:", err)
	}
	defer rc.Close()

	count := 0
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal("error on read logs:", err)
	}

	if count != expectedLines {
		t.Errorf("expected %d, got %d", expectedLines, count)
	}
}

func TestFakeOperationsLogsFollow(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Name: "gopher@luizalabs.com"}
	app := &App{Name: "teresa"}

	fake.(*FakeOperations).Storage[app.Name] = app

	minimumLines := 1
	rc, err := fake.Logs(user, app.Name, int64(minimumLines), true)
	if err != nil {
		t.Fatal("error on get logs:", err)
	}

	count := 0
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {
		t.Fatal("error on read logs:", err)
	}

	if count < minimumLines {
		t.Errorf("expected more than %d, got %d", minimumLines, count)
	}
}

func TestFakeOperationsLogsErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Email: "bad-user@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage[app.Name] = app

	if _, err := fake.Logs(user, app.Name, 1, false); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestFakeOperationsLogsErrNotFound(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Email: "gopher-user@luizalabs.com"}
	app := &App{Name: "teresa"}

	if _, err := fake.Logs(user, app.Name, 1, false); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFakeOperationsInfo(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Name: "gopher@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage[app.Name] = app

	info, err := fake.Info(user, app.Name)
	if err != nil {
		t.Fatal("error getting app info: ", err)
	}

	if info == nil {
		t.Fatal("expected a valid info, got nil")
	}
}

func TestFakeOperationsInfoErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Email: "bad-user@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage[app.Name] = app

	if _, err := fake.Info(user, app.Name); teresa_errors.Get(err) != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", teresa_errors.Get(err))
	}
}

func TestFakeOperationsInfoErrNotFound(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Name: "gopher@luizalabs.com"}

	if _, err := fake.Info(user, "teresa"); teresa_errors.Get(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", teresa_errors.Get(err))
	}
}

func TestFakeOperationsList(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Name: "gopher@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage[app.Name] = app

	apps, err := fake.List(user)
	if err != nil {
		t.Fatal("error getting app info: ", err)
	}

	if len(apps) != 1 {
		t.Fatal("expected a valid app, got nil")
	}

	if apps[0].Name != app.Name {
		t.Errorf("expected test, got %s", apps[0].Name)
	}
}

func TestFakeOperationsTeamName(t *testing.T) {
	fake := NewFakeOperations()
	teamName, err := fake.TeamName("teresa")
	if err != nil {
		t.Errorf("got error on get TeamName:", err)
	}
	if teamName != "luizalabs" {
		t.Errorf("expected luizalabs, got %s", teamName)
	}
}

func TestFakeOperationsMeta(t *testing.T) {
	fake := NewFakeOperations()
	a, err := fake.Get("teresa")
	if err != nil {
		t.Errorf("got error on get app Meta:", err)
	}
	if a.Name != "teresa" {
		t.Errorf("expected teresa, got %s", a.Name)
	}
}

func TestFakeHasPermission(t *testing.T) {
	var testCases = []struct {
		email    string
		expected bool
	}{
		{"gopher@luizalabs.com", true},
		{"bad-user@luizalabs.com", false},
	}
	fake := NewFakeOperations()

	for _, tc := range testCases {
		actual := fake.HasPermission(&database.User{Email: tc.email}, "teresa")
		if actual != tc.expected {
			t.Errorf("expected %v, got %v", tc.expected, actual)
		}
	}
}

func TestFakeOperationsSetEnv(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Name: "gopher@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage[app.Name] = app

	if err := fake.SetEnv(user, app.Name, nil); err != nil {
		t.Fatal("error setting app env: ", err)
	}
}

func TestFakeOperationsSetEnvErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Email: "bad-user@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage[app.Name] = app

	if err := fake.SetEnv(user, app.Name, nil); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestFakeOperationsSetEnvErrNotFound(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Name: "gopher@luizalabs.com"}

	if err := fake.SetEnv(user, "teresa", nil); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFakeOperationsUnsetEnv(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Name: "gopher@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage[app.Name] = app

	if err := fake.UnsetEnv(user, app.Name, nil); err != nil {
		t.Fatal("error unsetting app env: ", err)
	}
}

func TestFakeOperationsUnsetEnvErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Email: "bad-user@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage[app.Name] = app

	if err := fake.UnsetEnv(user, app.Name, nil); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestFakeOperationsUnsetEnvErrNotFound(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Name: "gopher@luizalabs.com"}

	if err := fake.UnsetEnv(user, "teresa", nil); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFakeOperationsSetAutoscale(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Name: "gopher@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage[app.Name] = app

	req := newAutoscaleRequest("teresa")
	as := newAutoscale(req)

	if err := fake.SetAutoscale(user, app.Name, as); err != nil {
		t.Fatal("error on SetautoScale: ", err)
	}
}

func TestFakeOperationsSetAutoscalePermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Email: "bad-user@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage[app.Name] = app

	if err := fake.SetAutoscale(user, app.Name, nil); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestFakeOperationsSetAutoscaletErrNotFound(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Name: "gopher@luizalabs.com"}

	if err := fake.SetAutoscale(user, "teresa", nil); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFakeOperationsCheckPermAndGet(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Name: "gopher@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage[app.Name] = app

	if _, err := fake.CheckPermAndGet(user, app.Name); err != nil {
		t.Fatal("error CheckPermAndGet: ", err)
	}
}

func TestFakeOperationsCheckPermAndGetPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Email: "bad-user@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage[app.Name] = app

	if _, err := fake.CheckPermAndGet(user, app.Name); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestFakeOperationsCheckPermAndGetErrNotFound(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Name: "gopher@luizalabs.com"}

	if _, err := fake.CheckPermAndGet(user, "app"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFakeOperationsSaveApp(t *testing.T) {
	fake := NewFakeOperations()
	app := &App{Name: "teresa"}

	if err := fake.SaveApp(app, "gopher@luizalabs.com"); err != nil {
		t.Fatal("error SaveApp: ", err)
	}
}

func TestFakeOperationsSaveAppErrNotFound(t *testing.T) {
	fake := NewFakeOperations()
	app := &App{Name: "bad-app"}

	if err := fake.SaveApp(app, "gopher@luizalabs.com"); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
