package user

import (
	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	userpb "github.com/luizalabs/teresa-api/pkg/protobuf"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

type Service struct {
	ops Operations
}

func (s *Service) Login(ctx context.Context, request *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	token, err := s.ops.Login(request.Email, request.Password)
	if err != nil {
		return nil, auth.ErrPermissionDenied
	}
	return &userpb.LoginResponse{Token: token}, nil
}

func (s *Service) RegisterService(grpcServer *grpc.Server) {
	userpb.RegisterUserServer(grpcServer, s)
}

func NewService(ops Operations) *Service {
	return &Service{ops: ops}
}
