package client

import (
	"crypto/tls"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type tokenAuth struct {
	token string
}

func (t *tokenAuth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"token": t.token}, nil
}

func (*tokenAuth) RequireTransportSecurity() bool { return true }

func New(cfg ClusterConfig) (*grpc.ClientConn, error) {
	tlsConfig := new(tls.Config)
	creds := credentials.NewTLS(tlsConfig)
	return grpc.Dial(
		cfg.Server,
		grpc.WithPerRPCCredentials(&tokenAuth{cfg.Token}),
		grpc.WithTransportCredentials(creds),
	)
}
