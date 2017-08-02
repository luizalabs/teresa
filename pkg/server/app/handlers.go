package app

import (
	"bufio"

	log "github.com/Sirupsen/logrus"

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

<<<<<<< ca65c361ae0acaeb531bc67b5f6b29b4f8aa7186
func (s *Service) Info(ctx context.Context, req *appb.InfoRequest) (*appb.InfoResponse, error) {
	user := ctx.Value("user").(*storage.User)

	info, err := s.ops.Info(user, req.Name)
	if err != nil {
		return nil, err
	}

	return newInfoResponse(info), nil
}

func (s *Service) SetEnv(ctx context.Context, req *appb.SetEnvRequest) (*appb.Empty, error) {
	user := ctx.Value("user").(*storage.User)
	evs := newEnvVars(req)

	if err := s.ops.SetEnv(user, req.Name, evs); err != nil {
		return nil, err
	}

	return &appb.Empty{}, nil
}

func (s *Service) UnsetEnv(ctx context.Context, req *appb.UnsetEnvRequest) (*appb.Empty, error) {
	user := ctx.Value("user").(*storage.User)

	if err := s.ops.UnsetEnv(user, req.Name, req.EnvVars); err != nil {
		return nil, err
	}

	return &appb.Empty{}, nil
}

func (s *Service) List(req *appb.Empty, stream appb.ListResponse) error {
	ctx := stream.Context()
	user := ctx.Value("user").(*storage.User)

	list, err := s.ops.List(user)
	if err != nil {
		log.Errorf("app list failed: %v", err)
		return nil, grpcErr(err)
	}

	return newListResponse(list), nil
}

func (s *Service) RegisterService(grpcServer *grpc.Server) {
	appb.RegisterAppServer(grpcServer, s)
}

func NewService(ops Operations) *Service {
	return &Service{ops: ops}
}
