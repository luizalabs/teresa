package user

import (
	"time"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	userpb "github.com/luizalabs/teresa/pkg/protobuf/user"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/database"
)

type Service struct {
	ops Operations
}

func (s *Service) Login(ctx context.Context, request *userpb.LoginRequest) (*userpb.LoginResponse, error) {
	exp := time.Hour * 24 * 15
	if request.ExpiresIn != 0 {
		exp = time.Duration(request.ExpiresIn)
	}
	token, err := s.ops.Login(request.Email, request.Password, exp)
	if err != nil {
		return nil, auth.ErrPermissionDenied
	}
	return &userpb.LoginResponse{Token: token}, nil
}

func (s *Service) SetPassword(ctx context.Context, request *userpb.SetPasswordRequest) (*userpb.Empty, error) {
	u := ctx.Value("user").(*database.User)
	if err := s.ops.SetPassword(u, request.Password, request.User); err != nil {
		return nil, err
	}
	return &userpb.Empty{}, nil
}

func (s *Service) Delete(ctx context.Context, request *userpb.DeleteRequest) (*userpb.Empty, error) {
	u := ctx.Value("user").(*database.User)
	if !u.IsAdmin {
		return nil, auth.ErrPermissionDenied
	}
	if err := s.ops.Delete(request.Email); err != nil {
		return nil, err
	}
	return &userpb.Empty{}, nil
}

func (s *Service) Create(ctx context.Context, request *userpb.CreateRequest) (*userpb.Empty, error) {
	u := ctx.Value("user").(*database.User)
	if !u.IsAdmin {
		return nil, auth.ErrPermissionDenied
	}
	err := s.ops.Create(
		request.Name,
		request.Email,
		request.Password,
		request.Admin,
	)
	if err != nil {
		return nil, err
	}
	return &userpb.Empty{}, nil
}

func (s *Service) RegisterService(grpcServer *grpc.Server) {
	userpb.RegisterUserServer(grpcServer, s)
}

func NewService(ops Operations) *Service {
	return &Service{ops: ops}
}
