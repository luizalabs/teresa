package server

import (
	"net"
	"strings"

	"golang.org/x/net/context"

	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa-api/pkg/server/auth"
	"github.com/luizalabs/teresa-api/pkg/server/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

type Options struct {
	Port    string
	TLSCert string
	TLSKey  string
	Auth    auth.Auth
	DB      *gorm.DB
}

type Server struct {
	listener   net.Listener
	grpcServer *grpc.Server
	auth       auth.Auth
}

func unaryInterceptor(a auth.Auth) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if strings.HasSuffix(info.FullMethod, "Login") {
			return handler(ctx, req)
		}

		email, err := authorize(ctx, a)
		if err != nil {
			return nil, err
		}

		ctx = context.WithValue(ctx, "email", email)
		return handler(ctx, req)
	}
}

func authorize(ctx context.Context, a auth.Auth) (string, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return "", auth.ErrPermissionDenied
	}
	if len(md["token"]) < 1 || md["token"][0] == "" {
		return "", auth.ErrPermissionDenied
	}
	return a.ValidateToken(md["token"][0])
}

func (s *Server) Run() error {
	return s.grpcServer.Serve(s.listener)
}

func New(opt Options) (*Server, error) {
	l, err := net.Listen("tcp", opt.Port)
	if err != nil {
		return nil, err
	}

	creds, err := credentials.NewServerTLSFromFile(opt.TLSCert, opt.TLSKey)
	if err != nil {
		return nil, err
	}

	s := grpc.NewServer(
		grpc.Creds(creds),
		grpc.UnaryInterceptor(unaryInterceptor(opt.Auth)),
	)

	us := user.NewService(user.NewDatabaseOperations(opt.DB, opt.Auth))
	us.RegisterService(s)

	return &Server{listener: l, grpcServer: s, auth: opt.Auth}, nil
}
