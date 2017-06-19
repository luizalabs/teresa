package app

import (
	"bufio"

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

func (s *Service) Logs(req *appb.LogsRequest, stream appb.App_LogsServer) error {
	ctx := stream.Context()
	user := ctx.Value("user").(*storage.User)

	rc, err := s.ops.Logs(user, req.Name, req.Lines, req.Follow)
	if err != nil {
		return err
	}
	defer rc.Close()

	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		if err := stream.Send(&appb.LogsResponse{Text: scanner.Text()}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Info(ctx context.Context, req *appb.InfoRequest) (*appb.InfoResponse, error) {
	user := ctx.Value("user").(*storage.User)

	info, err := s.ops.Info(user, req.Name)
	if err != nil {
		return nil, err
	}

	return newInfoResponse(info), nil
}

func (s *Service) RegisterService(grpcServer *grpc.Server) {
	appb.RegisterAppServer(grpcServer, s)
}

func NewService(ops Operations) *Service {
	return &Service{ops: ops}
}
