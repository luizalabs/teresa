package exec

import (
	"time"

	"github.com/luizalabs/teresa/pkg/goutil"
	execpb "github.com/luizalabs/teresa/pkg/protobuf/exec"
	"github.com/luizalabs/teresa/pkg/server/database"
	"google.golang.org/grpc"
)

const keepAliveMessage = "\u200B" // Zero width space

type Service struct {
	ops              Operations
	keepAliveTimeout time.Duration
}

func (s *Service) Command(req *execpb.CommandRequest, stream execpb.Exec_CommandServer) error {
	ctx := stream.Context()
	u := ctx.Value("user").(*database.User)

	rc, errChan := s.ops.RunCommand(ctx, u, req.AppName, req.Command...)
	if rc == nil {
		return <-errChan
	}
	defer rc.Close()

	cmdMsgs, cmdErrCh := goutil.LineGenerator(rc)
	var msg string
	for {
		select {
		case <-time.After(s.keepAliveTimeout):
			msg = keepAliveMessage
		case err := <-errChan:
			return err
		case err := <-cmdErrCh:
			return err
		case m, ok := <-cmdMsgs:
			if !ok {
				return nil
			}
			msg = m
		}

		if err := stream.Send(&execpb.CommandResponse{Text: msg + "\n"}); err != nil {
			return err
		}
	}
}

func (s *Service) RegisterService(grpcServer *grpc.Server) {
	execpb.RegisterExecServer(grpcServer, s)
}

func NewService(ops Operations, keepAliveTimeout time.Duration) *Service {
	return &Service{ops: ops, keepAliveTimeout: keepAliveTimeout}
}
