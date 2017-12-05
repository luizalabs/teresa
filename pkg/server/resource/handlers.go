package resource

import (
	respb "github.com/luizalabs/teresa/pkg/protobuf/resource"
	"github.com/luizalabs/teresa/pkg/server/database"

	context "golang.org/x/net/context"

	"google.golang.org/grpc"
)

type Service struct {
	ops Operations
}

func (svc *Service) Create(ctx context.Context, req *respb.CreateRequest) (*respb.CreateResponse, error) {
	user := ctx.Value("user").(*database.User)
	res := newResource(req)

	text, err := svc.ops.Create(user, res)
	if err != nil {
		return nil, err
	}

	return &respb.CreateResponse{text}, nil
}

func (svc *Service) Delete(ctx context.Context, req *respb.DeleteRequest) (*respb.Empty, error) {
	user := ctx.Value("user").(*database.User)

	if err := svc.ops.Delete(user, req.Name); err != nil {
		return nil, err
	}

	return &respb.Empty{}, nil
}

func (svc *Service) RegisterService(grpcServer *grpc.Server) {
	respb.RegisterResourceServer(grpcServer, svc)
}

func NewService(ops Operations) *Service {
	return &Service{ops: ops}
}
