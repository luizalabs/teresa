package team

import (
	context "golang.org/x/net/context"

	"github.com/luizalabs/teresa-api/models/storage"
	teampb "github.com/luizalabs/teresa-api/pkg/protobuf/team"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	"google.golang.org/grpc"
)

type Service struct {
	ops Operations
}

func (s *Service) Create(ctx context.Context, request *teampb.CreateRequest) (*teampb.Empty, error) {
	u := ctx.Value("user").(*storage.User)
	if !u.IsAdmin {
		return nil, auth.ErrPermissionDenied
	}
	if err := s.ops.Create(request.Name, request.Email, request.Url); err != nil {
		return nil, err
	}
	return &teampb.Empty{}, nil
}

func (s *Service) RegisterService(grpcServer *grpc.Server) {
	teampb.RegisterTeamServer(grpcServer, s)
}

func NewService(ops Operations) *Service {
	return &Service{ops: ops}
}
