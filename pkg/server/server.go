package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/auth"
	"github.com/luizalabs/teresa/pkg/server/build"
	"github.com/luizalabs/teresa/pkg/server/cloudprovider"
	"github.com/luizalabs/teresa/pkg/server/deploy"
	"github.com/luizalabs/teresa/pkg/server/exec"
	"github.com/luizalabs/teresa/pkg/server/healthcheck"
	"github.com/luizalabs/teresa/pkg/server/k8s"
	"github.com/luizalabs/teresa/pkg/server/service"
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
	K8s       *k8s.Client
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

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)
	defer close(exitChan)

	gChan := make(chan error)
	go func() { gChan <- g.Wait() }()

	select {
	case err := <-gChan:
		return err
	case <-exitChan:
		s.grpcServer.GracefulStop()
		s.hcServer.GracefulStop()
		return nil
	}
}

func createServerOps(opt Options, uOps user.Operations) []grpc.ServerOption {
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

func registerServices(s *grpc.Server, opt Options, uOps user.Operations) error {
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

	execDefaults := &exec.Defaults{
		RunnerImage:  opt.DeployOpt.SlugRunnerImage,
		StoreImage:   opt.DeployOpt.SlugStoreImage,
		LimitsCPU:    opt.DeployOpt.BuildLimitCPU,
		LimitsMemory: opt.DeployOpt.BuildLimitMemory,
	}
	execOps := exec.NewOperations(appOps, opt.K8s, opt.Storage, execDefaults)
	e := exec.NewService(execOps)
	e.RegisterService(s)

	buildOpts := &build.Options{
		SlugBuilderImage: opt.DeployOpt.SlugBuilderImage,
		BuildLimitCPU:    opt.DeployOpt.BuildLimitCPU,
		BuildLimitMemory: opt.DeployOpt.BuildLimitMemory,
	}
	bOps := build.NewBuildOperations(opt.Storage, execOps, buildOpts)

	dOps := deploy.NewDeployOperations(appOps, opt.K8s, opt.Storage, execOps, bOps, opt.DeployOpt)
	d := deploy.NewService(dOps, opt.DeployOpt)
	d.RegisterService(s)

	cpOps, err := cloudprovider.NewOperations(opt.K8s)
	if err != nil {
		return err
	}

	svcOps := service.NewOperations(appOps, cpOps, opt.K8s)
	svc := service.NewService(svcOps)
	svc.RegisterService(s)
	return nil
}

func New(opt Options) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", opt.Port))
	if err != nil {
		return nil, err
	}

	uOps := user.NewDatabaseOperations(opt.DB, opt.Auth)
	sOpts := createServerOps(opt, uOps)
	s := grpc.NewServer(sOpts...)
	if err := registerServices(s, opt, uOps); err != nil {
		return nil, err
	}

	hcServer := healthcheck.New(opt.K8s, opt.DB)
	return &Server{listener: l, grpcServer: s, hcServer: hcServer, opt: &opt}, nil
}
