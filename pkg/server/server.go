package server

import (
	"crypto/tls"
	"fmt"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa-api/pkg/server/app"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	"github.com/luizalabs/teresa-api/pkg/server/deploy"
	"github.com/luizalabs/teresa-api/pkg/server/k8s"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
	"github.com/luizalabs/teresa-api/pkg/server/team"
	"github.com/luizalabs/teresa-api/pkg/server/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Options struct {
	Port      string
	TLSCert   *tls.Certificate
	Auth      auth.Auth
	DB        *gorm.DB
	Storage   st.Storage
	K8s       k8s.Client
	DeployOpt *deploy.Options
}

type Server struct {
	listener   net.Listener
	grpcServer *grpc.Server
}

func (s *Server) Run() error {
	return s.grpcServer.Serve(s.listener)
}

func New(opt Options) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", opt.Port))
	if err != nil {
		return nil, err
	}

	uOps := user.NewDatabaseOperations(opt.DB, opt.Auth)
	recOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(recFunc),
	}
	sOpts := []grpc.ServerOption{
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			loginUnaryInterceptor(opt.Auth, uOps),
			logUnaryInterceptor,
			grpc_recovery.UnaryServerInterceptor(recOpts...),
		)),
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			loginStreamInterceptor(opt.Auth, uOps),
			logStreamInterceptor,
			grpc_recovery.StreamServerInterceptor(recOpts...),
		)),
	}
	if opt.TLSCert != nil {
		creds := credentials.NewServerTLSFromCert(opt.TLSCert)
		sOpts = append(sOpts, grpc.Creds(creds))
	}

	s := grpc.NewServer(sOpts...)

	us := user.NewService(uOps)
	us.RegisterService(s)

	tOps := team.NewDatabaseOperations(opt.DB, uOps)
	t := team.NewService(tOps)
	t.RegisterService(s)

	appOps := app.NewOperations(tOps, opt.K8s, opt.Storage)
	a := app.NewService(appOps)
	a.RegisterService(s)

	dOps := deploy.NewDeployOperations(appOps, opt.K8s, opt.Storage)
	d := deploy.NewService(dOps, opt.DeployOpt)
	d.RegisterService(s)

	return &Server{listener: l, grpcServer: s}, nil
}
