package build

import (
	"testing"
	"time"

	bpb "github.com/luizalabs/teresa/pkg/protobuf/build"
	"github.com/luizalabs/teresa/pkg/server/database"
	context "golang.org/x/net/context"
)

func TestListSuccess(t *testing.T) {
	fake := NewFakeOperations()

	srv := NewService(fake, time.Second*1)
	req := &bpb.ListRequest{AppName: "teresa"}
	u := &database.User{Email: "gopher@luizalabs.com"}

	ctx := context.WithValue(context.Background(), "user", u)
	resp, err := srv.List(ctx, req)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if len(resp.Builds) != 2 { // see fake.go
		t.Errorf(
			"expected [v1.0.0 v1.1.0-rc1], got [%s %s]",
			resp.Builds[0].Name,
			resp.Builds[1].Name,
		)
	}

}

func TestDeleteSucess(t *testing.T) {
	fake := NewFakeOperations()

	srv := NewService(fake, time.Second*1)
	req := &bpb.DeleteRequest{AppName: "teresa", Name: "fake"}
	u := &database.User{Email: "gopher@luizalabs.com"}

	ctx := context.WithValue(context.Background(), "user", u)
	if _, err := srv.Delete(ctx, req); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
