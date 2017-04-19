package user

import (
	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/luizalabs/teresa-api/pkg/common"
	userpb "github.com/luizalabs/teresa-api/pkg/protobuf"
)

type Server struct {
	ops Operations
}

func (s *Server) Login(ctx context.Context, request *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	token, err := s.ops.Login(request.Email, request.Password)
	if err != nil {
		return nil, common.ErrPermissionDenied
	}
	return &userpb.LoginResponse{Token: token}, nil
}

func (s *Server) RegisterServer(grpcServer *grpc.Server) {
	userpb.RegisterUserServer(grpcServer, s)
}

func NewServer(ops Operations) *Server {
	return &Server{ops: ops}
}
