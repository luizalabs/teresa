package resource

import (
	"testing"

	respb "github.com/luizalabs/teresa/pkg/protobuf/resource"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"

	context "golang.org/x/net/context"
)

func TestCreateSuccess(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Email: "gopher@luizalabs.com"}
	svc := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := svc.Create(ctx, &respb.CreateRequest{Name: "teresa"}); err != nil {
		t.Error("got error on create:", err)
	}
}

func TestCreateErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	svc := NewService(fake)
	user := &database.User{Email: "bad-user@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := svc.Create(ctx, &respb.CreateRequest{Name: "teresa"}); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestCreateErrAlreadyExists(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Email: "gopher@luizalabs.com"}
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &Resource{Name: name}
	svc := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := svc.Create(ctx, &respb.CreateRequest{Name: name}); err != ErrAlreadyExists {
		t.Errorf("expected ErrAlreadyExists, got %s", err)
	}
}

func TestDeleteSuccess(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Email: "gopher@luizalabs.com"}
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = &Resource{Name: name}
	svc := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := svc.Delete(ctx, &respb.DeleteRequest{Name: name}); err != nil {
		t.Error("got error on delete:", err)
	}
}

func TestDeleteErrPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	svc := NewService(fake)
	user := &database.User{Email: "bad-user@luizalabs.com"}
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := svc.Delete(ctx, &respb.DeleteRequest{Name: "teresa"}); err != auth.ErrPermissionDenied {
		t.Errorf("expected ErrPermissionDenied, got %s", err)
	}
}

func TestDeleteErrNotFound(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Email: "gopher@luizalabs.com"}
	svc := NewService(fake)
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := svc.Delete(ctx, &respb.DeleteRequest{Name: "teresa"}); err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %s", err)
	}
}
