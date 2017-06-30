package app

import (
	"bufio"
	"testing"

	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

func TestFakeOperationsCreate(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	user := &storage.User{Name: "gopher@luizalabs.com"}
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
	user := &storage.User{Email: "gopher@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage["teresa"] = app

	if err := fake.Create(user, app); err != ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestFakeOperationsCreateErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	user := &storage.User{Email: "bad-user@luizalabs.com"}
	app := &App{Name: "teresa"}

	if err := fake.Create(user, app); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestFakeOperationsLogs(t *testing.T) {
	fake := NewFakeOperations()
	user := &storage.User{Name: "gopher@luizalabs.com"}
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
	user := &storage.User{Name: "gopher@luizalabs.com"}
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
	user := &storage.User{Email: "bad-user@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage[app.Name] = app

	if _, err := fake.Logs(user, app.Name, 1, false); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestFakeOperationsLogsErrNotFound(t *testing.T) {
	fake := NewFakeOperations()
	user := &storage.User{Email: "gopher-user@luizalabs.com"}
	app := &App{Name: "teresa"}

	if _, err := fake.Logs(user, app.Name, 1, false); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFakeOperationsInfo(t *testing.T) {
	fake := NewFakeOperations()
	user := &storage.User{Name: "gopher@luizalabs.com"}
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
	user := &storage.User{Email: "bad-user@luizalabs.com"}
	app := &App{Name: "teresa"}
	fake.(*FakeOperations).Storage[app.Name] = app

	if _, err := fake.Info(user, app.Name); grpcErr(err) != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", grpcErr(err))
	}
}

func TestFakeOperationsInfoErrNotFound(t *testing.T) {
	fake := NewFakeOperations()
	user := &storage.User{Name: "gopher@luizalabs.com"}

	if _, err := fake.Info(user, "teresa"); grpcErr(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", grpcErr(err))
	}
}
