package exec

import (
	"testing"

	context "golang.org/x/net/context"

	execpb "github.com/luizalabs/teresa/pkg/protobuf/exec"
	"github.com/luizalabs/teresa/pkg/server/database"
)

type streamWrapper struct {
	execpb.Exec_CommandServer
	ctx context.Context
}

func (sw *streamWrapper) Context() context.Context {
	return sw.ctx
}

func TestCommand(t *testing.T) {
	s := NewService(NewFakeOperations())

	user := &database.User{}
	ctx := context.WithValue(context.Background(), "user", user)
	req := &execpb.CommandRequest{Name: "teresa", Command: []string{"ls"}}

	wrap := &streamWrapper{ctx: ctx}
	if err := s.Command(req, wrap); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
