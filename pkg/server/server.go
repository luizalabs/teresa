package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"golang.org/x/net/context"

	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa-api/models/storage"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	"github.com/luizalabs/teresa-api/pkg/server/team"
	"github.com/luizalabs/teresa-api/pkg/server/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type Options struct {
	Port    string
	TLSCert *tls.Certificate
	Auth    auth.Auth
	DB      *gorm.DB
}

type Server struct {
	listener   net.Listener
	grpcServer *grpc.Server
}

func unaryInterceptor(a auth.Auth, uOps user.Operations) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if strings.HasSuffix(info.FullMethod, "Login") {
			return handler(ctx, req)
		}

		user, err := authorize(ctx, a, uOps)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, "user", user)
		return handler(ctx, req)
	}
}

func authorize(ctx context.Context, a auth.Auth, uOps user.Operations) (*storage.User, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return nil, auth.ErrPermissionDenied
	}
	if len(md["token"]) < 1 || md["token"][0] == "" {
		return nil, auth.ErrPermissionDenied
	}
	email, err := a.ValidateToken(md["token"][0])
	if err != nil {
		return nil, err
	}
	return uOps.GetUser(email)
}

func (s *Server) Run() error {
	return s.grpcServer.Serve(s.listener)
}

func New(opt Options) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%s", opt.Port))
	if err != nil {
		return nil, err
	}

	creds := credentials.NewServerTLSFromCert(opt.TLSCert)

	uOps := user.NewDatabaseOperations(opt.DB, opt.Auth)
	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(unaryInterceptor(opt.Auth, uOps)),
	)

	us := user.NewService(uOps)
	us.RegisterService(s)

	tOps := team.NewDatabaseOperations(opt.DB)
	t := team.NewService(tOps)
	t.RegisterService(s)

	return &Server{listener: l, grpcServer: s}, nil
}
