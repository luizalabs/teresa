package server

import (
	"crypto/rsa"
	"net"

	"github.com/jinzhu/gorm"
	"github.com/luizalabs/teresa-api/pkg/server/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Options struct {
	Port          string
	TLSCert       string
	TLSKey        string
	RSAPrivateKey *rsa.PrivateKey
	DB            *gorm.DB
}

type Server struct {
	listener   net.Listener
	grpcServer *grpc.Server
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

	s := grpc.NewServer(grpc.Creds(creds))

	cs := user.NewServer(user.NewDatabaseOperations(opt.DB, opt.RSAPrivateKey))
	cs.RegisterServer(s)

	return &Server{listener: l, grpcServer: s}, nil
}
