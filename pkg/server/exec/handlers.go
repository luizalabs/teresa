package exec

import (
	"github.com/luizalabs/teresa/pkg/goutil"
	execpb "github.com/luizalabs/teresa/pkg/protobuf/exec"
	"github.com/luizalabs/teresa/pkg/server/database"
	"google.golang.org/grpc"
)

type Service struct {
	ops Operations
}

func (s *Service) Command(req *execpb.CommandRequest, stream execpb.Exec_CommandServer) error {
	ctx := stream.Context()
	u := ctx.Value("user").(*database.User)

	rc, errChan := s.ops.Command(u, req.Name, req.Command...)
	if rc == nil {
		return <-errChan
	}
	defer rc.Close()

	cmdMsgs := goutil.ChannelFromReader(rc, true)
	var msg string
	for {
		select {
		case err := <-errChan:
			return err
		case m, ok := <-cmdMsgs:
			if !ok {
				return nil
			}
			msg = m
		}

		if err := stream.Send(&execpb.CommandResponse{Text: msg}); err != nil {
			return err
		}
	}
}

func (s *Service) RegisterService(grpcServer *grpc.Server) {
	execpb.RegisterExecServer(grpcServer, s)
}

func NewService(ops Operations) *Service {
	return &Service{ops: ops}
}
