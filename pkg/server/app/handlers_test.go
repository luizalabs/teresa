package app

import (
	"bytes"

	context "golang.org/x/net/context"

	"testing"

	appb "github.com/luizalabs/teresa/pkg/protobuf/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
	"github.com/luizalabs/teresa/pkg/server/teresa_errors"
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
	user := &database.User{Email: "gopher@luizalabs.com"}
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
	user := &database.User{Email: "bad-user@luizalabs.com"}
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
	user := &database.User{Email: "gopher@luizalabs.com"}
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
	user := &database.User{Email: "gopher@luizalabs.com"}

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
	user := &database.User{Email: "gopher@luizalabs.com"}

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
	user := &database.User{Email: "bad-user@luizalabs.com"}

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
	user := &database.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.Info(ctx, &appb.InfoRequest{Name: name}); err != nil {
		t.Error("Got error on info: ", err)
	}
}

func TestInfoAppNotFound(t *testing.T) {
	s := NewService(NewFakeOperations())
	user := &database.User{Email: "gopher@luizalabs.com"}
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
	user := &database.User{Email: "bad-user@luizalabs.com"}
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
	user := &database.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.List(ctx, &appb.Empty{}); err != nil {
		t.Error("Got error on list: ", err)
	}
}

func TestSetEnvSuccess(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &database.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.SetEnv(ctx, &appb.SetEnvRequest{Name: name}); err != nil {
		t.Error("Got error on set env: ", err)
	}
}

func TestSetEnvAppNotFound(t *testing.T) {
	s := NewService(NewFakeOperations())
	user := &database.User{Email: "gopher@luizalabs.com"}
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
	user := &database.User{Email: "bad-user@luizalabs.com"}
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
	user := &database.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.UnsetEnv(ctx, &appb.UnsetEnvRequest{Name: name}); err != nil {
		t.Error("Got error on unset env: ", err)
	}
}

func TestUnsetEnvAppNotFound(t *testing.T) {
	s := NewService(NewFakeOperations())
	user := &database.User{Email: "gopher@luizalabs.com"}
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
	user := &database.User{Email: "bad-user@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := s.UnsetEnv(ctx, &appb.UnsetEnvRequest{Name: name}); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func newAutoscaleRequest(name string) *appb.SetAutoscaleRequest {
	as := &appb.SetAutoscaleRequest_Autoscale{
		Min:                  1,
		Max:                  2,
		CpuTargetUtilization: 10,
	}

	return &appb.SetAutoscaleRequest{
		Name:      name,
		Autoscale: as,
	}
}

func TestSetAutoscaleSuccess(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &database.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	req := newAutoscaleRequest(name)
	if _, err := s.SetAutoscale(ctx, req); err != nil {
		t.Error("Got error on autoscale: ", err)
	}
}

func TestSetAutoscaleAppNotFound(t *testing.T) {
	name := "teresa"
	s := NewService(NewFakeOperations())
	user := &database.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	req := newAutoscaleRequest(name)
	if _, err := s.SetAutoscale(ctx, req); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSetAutoscalePermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &database.User{Email: "bad-user@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	req := newAutoscaleRequest(name)

	if _, err := s.SetAutoscale(ctx, req); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestDeleteSuccess(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &database.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	req := &appb.DeleteRequest{Name: name}
	if _, err := s.Delete(ctx, req); err != nil {
		t.Error("Got error on delete: ", err)
	}
}

func TestDeleteAppNotFound(t *testing.T) {
	name := "teresa"
	s := NewService(NewFakeOperations())
	user := &database.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	req := &appb.DeleteRequest{Name: name}
	if _, err := s.Delete(ctx, req); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeletePermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &database.User{Email: "bad-user@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	req := &appb.DeleteRequest{Name: name}

	if _, err := s.Delete(ctx, req); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestSetReplicasSuccess(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &database.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	req := &appb.SetReplicasRequest{Name: name}
	if _, err := s.SetReplicas(ctx, req); err != nil {
		t.Error("got error on setReplicas: ", err)
	}
}

func TestSetReplicasAppNotFound(t *testing.T) {
	name := "teresa"
	s := NewService(NewFakeOperations())
	user := &database.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	req := &appb.SetReplicasRequest{Name: name}
	if _, err := s.SetReplicas(ctx, req); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestSetReplicasPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &database.User{Email: "bad-user@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	req := &appb.SetReplicasRequest{Name: name}

	if _, err := s.SetReplicas(ctx, req); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestDeletePodsSuccess(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &database.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)
	req := &appb.DeletePodsRequest{Name: name, PodsNames: []string{"pod1", "pod2"}}

	if _, err := s.DeletePods(ctx, req); err != nil {
		t.Error("got error on delete pods:", err)
	}
}

func TestDeletePodsErrNotFound(t *testing.T) {
	s := NewService(NewFakeOperations())
	user := &database.User{Email: "gopher@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)
	req := &appb.DeletePodsRequest{Name: "teresa", PodsNames: []string{"pod1", "pod2"}}

	if _, err := s.DeletePods(ctx, req); teresa_errors.Get(err) != ErrNotFound {
		t.Errorf("expected %v, got %v", ErrNotFound, teresa_errors.Get(err))
	}
}

func TestDeletePodsErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &App{Name: name}
	s := NewService(fake)
	user := &database.User{Email: "bad-user@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)
	req := &appb.DeletePodsRequest{Name: name, PodsNames: []string{"pod1", "pod2"}}

	if _, err := s.DeletePods(ctx, req); teresa_errors.Get(err) != auth.ErrPermissionDenied {
		t.Errorf("expected %v, got %v", auth.ErrPermissionDenied, teresa_errors.Get(err))
	}
}
