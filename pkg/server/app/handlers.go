package app

import (
	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/luizalabs/teresa-api/models/storage"
	appb "github.com/luizalabs/teresa-api/pkg/protobuf/app"
)

type Service struct {
	ops Operations
}

func (s *Service) Create(ctx context.Context, req *appb.CreateRequest) (*appb.Empty, error) {
	user := ctx.Value("user").(*storage.User)
	app := newApp(req)
	if err := s.ops.Create(user, app); err != nil {
		return nil, err
	}
	return &appb.Empty{}, nil
}

func (s *Service) RegisterService(grpcServer *grpc.Server) {
	appb.RegisterAppServer(grpcServer, s)
}

func NewService(ops Operations) *Service {
	return &Service{ops: ops}
}
