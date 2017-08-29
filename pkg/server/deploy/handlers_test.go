package deploy

import (
	"testing"

	context "golang.org/x/net/context"

	dpb "github.com/luizalabs/teresa/pkg/protobuf/deploy"
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
)

func TestListSuccess(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = true
	user := &database.User{Email: "gopher@luizalabs.com"}
	srv := NewService(fake, nil)
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := srv.List(ctx, &dpb.ListRequest{AppName: name}); err != nil {
		t.Error("got error on List: ", err)
	}
}

func TestListAppNotFound(t *testing.T) {
	fake := NewFakeOperations()
	user := &database.User{Email: "gopher@luizalabs.com"}
	srv := NewService(fake, nil)
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := srv.List(ctx, &dpb.ListRequest{AppName: "teresa"}); err != app.ErrNotFound {
		t.Errorf("expected app.ErrNotFound, got %s", err)
	}
}

func TestListPermissionDenied(t *testing.T) {
	fake := NewFakeOperations()
	name := "teresa"
	fake.(*FakeOperations).Storage[name] = true
	user := &database.User{Email: "bad-user@luizalabs.com"}
	srv := NewService(fake, nil)
	ctx := context.WithValue(context.Background(), "user", user)

	if _, err := srv.List(ctx, &dpb.ListRequest{AppName: name}); err != auth.ErrPermissionDenied {
		t.Errorf("expected auth.ErrPermissionDenied, got %s", err)
	}
}
