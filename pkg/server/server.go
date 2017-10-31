package server

import (
	"crypto/tls"
	"fmt"
	"net"

	"golang.org/x/sync/errgroup"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/deploy"
	"github.com/luizalabs/teresa/pkg/server/healthcheck"
	"github.com/luizalabs/teresa/pkg/server/k8s"
	st "github.com/luizalabs/teresa/pkg/server/storage"
	"github.com/luizalabs/teresa/pkg/server/team"
	"github.com/luizalabs/teresa/pkg/server/user"
	"github.com/soheilhy/cmux"

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
	Debug     bool
}

type Server struct {
	listener   net.Listener
	grpcServer *grpc.Server
	hcServer   *healthcheck.Server
	opt        *Options
}

func (s *Server) Run() error {
	m := cmux.New(s.listener)

	grpcMatchers := []cmux.Matcher{cmux.HTTP2HeaderField("content-type", "application/grpc")}
	if s.opt.TLSCert != nil {
		grpcMatchers = append(grpcMatchers, cmux.TLS())
	}
	grpcListener := m.Match(grpcMatchers...)
	httpListener := m.Match(cmux.HTTP1Fast())

	g := new(errgroup.Group)
	g.Go(func() error { return s.grpcServer.Serve(grpcListener) })
	g.Go(func() error { return s.hcServer.Run(httpListener) })
	g.Go(func() error { return m.Serve() })

	return g.Wait()
}

func createSeverOps(opt Options, uOps user.Operations) []grpc.ServerOption {
	recOpts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(buildRecFunc(opt.Debug)),
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
	return sOpts
}

func registerServices(s *grpc.Server, opt Options, uOps user.Operations) {
	us := user.NewService(uOps)
	us.RegisterService(s)

	tOps := team.NewDatabaseOperations(opt.DB, uOps)
	t := team.NewService(tOps)
	t.RegisterService(s)

	appOps := app.NewOperations(tOps, opt.K8s, opt.Storage)
	a := app.NewService(appOps)
	a.RegisterService(s)

	// use appOps as teamExt to avoid circular import
	tOps.SetTeamExt(appOps)

	dOps := deploy.NewDeployOperations(appOps, opt.K8s, opt.Storage)
	d := deploy.NewService(dOps, opt.DeployOpt)
	d.RegisterService(s)
}

func New(opt Options) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", opt.Port))
	if err != nil {
		return nil, err
	}

	uOps := user.NewDatabaseOperations(opt.DB, opt.Auth)
	sOpts := createSeverOps(opt, uOps)
	s := grpc.NewServer(sOpts...)
	registerServices(s, opt, uOps)

	hcServer := healthcheck.New(opt.K8s, opt.DB)
	return &Server{listener: l, grpcServer: s, hcServer: hcServer, opt: &opt}, nil
}
