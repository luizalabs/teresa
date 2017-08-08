package client

import (
	"crypto/tls"
	"errors"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	defaultConnTimeout = 5 * time.Second
)

type tokenAuth struct {
	token string
}

func (t *tokenAuth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"token": t.token}, nil
}

func (*tokenAuth) RequireTransportSecurity() bool { return false }

func New(cfg ClusterConfig) (*grpc.ClientConn, error) {
	tlsConfig := new(tls.Config)

	opts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(&tokenAuth{cfg.Token}),
		grpc.WithBlock(),
		grpc.WithTimeout(defaultConnTimeout),
	}
	if cfg.UseTLS {
		if cfg.Insecure {
			tlsConfig.InsecureSkipVerify = true
		}
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.Dial(cfg.Server, opts...)
	if err != nil {
		return nil, errors.New(cfg.Server)
	}
	return conn, nil
}
