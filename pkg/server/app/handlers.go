package app

import (
	"time"

	context "golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/luizalabs/teresa/pkg/goutil"
	appb "github.com/luizalabs/teresa/pkg/protobuf/app"
	"github.com/luizalabs/teresa/pkg/server/database"
)

const (
	logSeparatorInterval = 30 * time.Second
	logSeparator         = "----- No logs\n"
)

type Service struct {
	ops Operations
}

func (s *Service) Create(ctx context.Context, req *appb.CreateRequest) (*appb.Empty, error) {
	user := ctx.Value("user").(*database.User)
	app := newApp(req)
	if err := s.ops.Create(user, app); err != nil {
		return nil, err
	}
	return &appb.Empty{}, nil
}

func (s *Service) Logs(req *appb.LogsRequest, stream appb.App_LogsServer) error {
	ctx := stream.Context()
	user := ctx.Value("user").(*database.User)
	opts := &LogOptions{
		Lines:     req.Lines,
		Follow:    req.Follow,
		PodName:   req.PodName,
		Previous:  req.Previous,
		Container: req.Container,
	}

	rc, err := s.ops.Logs(user, req.Name, opts)
	if err != nil {
		return err
	}
	defer rc.Close()

	chLogs := goutil.ChannelFromReader(rc, false)
	var line string

	for {
		select {
		case <-time.After(logSeparatorInterval):
			line = logSeparator
		case m, ok := <-chLogs:
			if !ok {
				return nil
			}
			line = m
		}

		if err := stream.Send(&appb.LogsResponse{Text: line}); err != nil {
			return err
		}
	}
}

func (s *Service) Info(ctx context.Context, req *appb.InfoRequest) (*appb.InfoResponse, error) {
	user := ctx.Value("user").(*database.User)

	info, err := s.ops.Info(user, req.Name)
	if err != nil {
		return nil, err
	}

	return newInfoResponse(info), nil
}

func (s *Service) SetEnv(ctx context.Context, req *appb.SetEnvRequest) (*appb.Empty, error) {
	user := ctx.Value("user").(*database.User)
	evs := newEnvVars(req)

	if err := s.ops.SetEnv(user, req.Name, evs); err != nil {
		return nil, err
	}

	return &appb.Empty{}, nil
}

func (s *Service) UnsetEnv(ctx context.Context, req *appb.UnsetEnvRequest) (*appb.Empty, error) {
	user := ctx.Value("user").(*database.User)

	if err := s.ops.UnsetEnv(user, req.Name, req.EnvVars); err != nil {
		return nil, err
	}

	return &appb.Empty{}, nil
}

func (s *Service) SetSecret(ctx context.Context, req *appb.SetEnvRequest) (*appb.Empty, error) {
	user := ctx.Value("user").(*database.User)
	evs := newEnvVars(req)

	if err := s.ops.SetSecret(user, req.Name, evs); err != nil {
		return nil, err
	}

	return &appb.Empty{}, nil
}

func (s *Service) UnsetSecret(ctx context.Context, req *appb.UnsetEnvRequest) (*appb.Empty, error) {
	user := ctx.Value("user").(*database.User)

	if err := s.ops.UnsetSecret(user, req.Name, req.EnvVars); err != nil {
		return nil, err
	}

	return &appb.Empty{}, nil
}

func (s *Service) List(ctx context.Context, _ *appb.Empty) (*appb.ListResponse, error) {
	user := ctx.Value("user").(*database.User)

	apps, err := s.ops.List(user)
	if err != nil {
		return nil, err
	}

	return newListResponse(apps), nil
}

func (s *Service) Delete(ctx context.Context, req *appb.DeleteRequest) (*appb.Empty, error) {
	user := ctx.Value("user").(*database.User)

	if err := s.ops.Delete(user, req.Name); err != nil {
		return nil, err
	}

	return &appb.Empty{}, nil
}

func (s *Service) SetAutoscale(ctx context.Context, req *appb.SetAutoscaleRequest) (*appb.Empty, error) {
	user := ctx.Value("user").(*database.User)
	as := newAutoscale(req)

	if err := s.ops.SetAutoscale(user, req.Name, as); err != nil {
		return nil, err
	}

	return &appb.Empty{}, nil
}

func (s *Service) SetReplicas(ctx context.Context, req *appb.SetReplicasRequest) (*appb.Empty, error) {
	user := ctx.Value("user").(*database.User)

	if err := s.ops.SetReplicas(user, req.Name, req.Replicas); err != nil {
		return nil, err
	}

	return &appb.Empty{}, nil
}

func (s *Service) DeletePods(ctx context.Context, req *appb.DeletePodsRequest) (*appb.Empty, error) {
	user := ctx.Value("user").(*database.User)

	if err := s.ops.DeletePods(user, req.Name, req.PodsNames); err != nil {
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
