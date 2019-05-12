package service

import (
	svcpb "github.com/luizalabs/teresa/pkg/protobuf/service"
	"github.com/luizalabs/teresa/pkg/server/database"

	context "golang.org/x/net/context"

	"google.golang.org/grpc"
)

type Service struct {
	ops Operations
}

func (svc *Service) EnableSSL(ctx context.Context, req *svcpb.EnableSSLRequest) (*svcpb.Empty, error) {
	user := ctx.Value("user").(*database.User)
	if err := svc.ops.EnableSSL(user, req.AppName, req.Cert, req.Only); err != nil {
		return nil, err
	}
	return &svcpb.Empty{}, nil
}

func (svc *Service) SetStaticIp(ctx context.Context, req *svcpb.SetStaticIpRequest) (*svcpb.Empty, error) {
	user := ctx.Value("user").(*database.User)
	if err := svc.ops.SetStaticIp(user, req.AppName, req.AddressName); err != nil {
		return nil, err
	}
	return &svcpb.Empty{}, nil
}

func (svc *Service) Info(ctx context.Context, req *svcpb.InfoRequest) (*svcpb.InfoResponse, error) {
	user := ctx.Value("user").(*database.User)
	info, err := svc.ops.Info(user, req.AppName)
	if err != nil {
		return nil, err
	}
	return newInfoResponse(info), nil
}

func (svc *Service) WhitelistSourceRanges(ctx context.Context, req *svcpb.WhitelistSourceRangesRequest) (*svcpb.Empty, error) {
	user := ctx.Value("user").(*database.User)
	if err := svc.ops.WhitelistSourceRanges(user, req.AppName, req.SourceRanges); err != nil {
		return nil, err
	}
	return &svcpb.Empty{}, nil
}

func (svc *Service) RegisterService(grpcServer *grpc.Server) {
	svcpb.RegisterServiceServer(grpcServer, svc)
}

func NewService(ops Operations) *Service {
	return &Service{ops: ops}
}
