package app

import (
	"bytes"

	context "golang.org/x/net/context"

	"testing"

	"github.com/luizalabs/teresa-api/models/storage"
	appb "github.com/luizalabs/teresa-api/pkg/protobuf/app"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	"github.com/luizalabs/teresa-api/pkg/server/teresa_errors"
)

type LogsStreamWrapper struct {
	appb.App_LogsServer
	ctx    context.Context
	buffer bytes.Buffer
}

func (lsw *LogsStreamWrapper) Context() context.Context {
	return lsw.ctx
}

func (lsw *LogsStreamWrapper) Send(msg *appb.LogsResponse) error {
	lsw.buffer.Write([]byte(msg.Text))
	return nil
}

func TestCreateSuccess(t *testing.T) {
	fake := NewFakeOperations()
	user := &storage.User{Email: "gopher@luizalabs.com"}
	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)

	_, err := s.Create(
		ctx,
		&appb.CreateRequest{Name: "teresa"},
	)
	if err != nil {
		t.Error("Got error on Create: ", err)
	}
}

func TestCreateErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	s := NewService(fake)
	user := &storage.User{Email: "bad-user@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	_, err := s.Create(
		ctx,
		&appb.CreateRequest{Name: "teresa"},
	)
	if err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestCreateErrAppAlreadyExists(t *testing.T) {
	fake := NewFakeOperations()
	user := &storage.User{Email: "gopher@luizalabs.com"}
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)

	_, err := s.Create(
		ctx,
		&appb.CreateRequest{Name: name},
	)
	if err != ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %s", err)
	}
}

func TestLogsSuccess(t *testing.T) {
	fake := NewFakeOperations()
	user := &storage.User{Email: "gopher@luizalabs.com"}

	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)

	ctx := context.WithValue(context.Background(), "user", user)
	req := &appb.LogsRequest{Name: name, Lines: 1, Follow: false}

	wrap := &LogsStreamWrapper{ctx: ctx}
	if err := s.Logs(req, wrap); err != nil {
		t.Error("error getting logs:", err)
	}
}

func TestLogsAppNotFound(t *testing.T) {
	fake := NewFakeOperations()
	user := &storage.User{Email: "gopher@luizalabs.com"}

	s := NewService(fake)

	ctx := context.WithValue(context.Background(), "user", user)
	req := &appb.LogsRequest{Name: "teresa", Lines: 1, Follow: false}

	wrap := &LogsStreamWrapper{ctx: ctx}
	if err := s.Logs(req, wrap); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestLogsPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	user := &storage.User{Email: "bad-user@luizalabs.com"}

	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)

	ctx := context.WithValue(context.Background(), "user", user)
	req := &appb.LogsRequest{Name: name, Lines: 1, Follow: false}

	wrap := &LogsStreamWrapper{ctx: ctx}
	if err := s.Logs(req, wrap); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestInfoSuccess(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &storage.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.Info(ctx, &appb.InfoRequest{Name: name}); err != nil {
		t.Error("Got error on info: ", err)
	}
}

func TestInfoAppNotFound(t *testing.T) {
	s := NewService(NewFakeOperations())
	user := &storage.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.Info(ctx, &appb.InfoRequest{Name: "teresa"}); teresa_errors.Get(err) != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestInfoPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &storage.User{Email: "bad-user@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.Info(ctx, &appb.InfoRequest{Name: name}); teresa_errors.Get(err) != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestListSuccess(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &storage.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.List(ctx, &appb.Empty{}); err != nil {
		t.Error("Got error on list: ", err)
	}
}

/*
func TestListAppNotFound(t *testing.T) {
	s := NewService(NewFakeOperations())
	user := &storage.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.List(ctx, &appb.Empty{}); teresa_errors.Get(err) != nil {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
*/

func TestSetEnvSuccess(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &storage.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.SetEnv(ctx, &appb.SetEnvRequest{Name: name}); err != nil {
		t.Error("Got error on set env: ", err)
	}
}

func TestSetEnvAppNotFound(t *testing.T) {
	s := NewService(NewFakeOperations())
	user := &storage.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.SetEnv(ctx, &appb.SetEnvRequest{Name: "teresa"}); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSetEnvPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &storage.User{Email: "bad-user@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.SetEnv(ctx, &appb.SetEnvRequest{Name: name}); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestUnsetEnvSuccess(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &storage.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.UnsetEnv(ctx, &appb.UnsetEnvRequest{Name: name}); err != nil {
		t.Error("Got error on unset env: ", err)
	}
}

func TestUnsetEnvAppNotFound(t *testing.T) {
	s := NewService(NewFakeOperations())
	user := &storage.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.UnsetEnv(ctx, &appb.UnsetEnvRequest{Name: "teresa"}); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUnsetEnvPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &storage.User{Email: "bad-user@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.UnsetEnv(ctx, &appb.UnsetEnvRequest{Name: name}); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}
