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

func (*tokenAuth) RequireTransportSecurity() bool { return false }

func New(cfg ClusterConfig, noTLS, tlsInsecure bool) (*grpc.ClientConn, error) {
	tlsConfig := new(tls.Config)

	opts := []grpc.DialOption{
		grpc.WithPerRPCCredentials(&tokenAuth{cfg.Token}),
	}
	if noTLS {
		opts = append(opts, grpc.WithInsecure())
	} else {
		if tlsInsecure {
			tlsConfig.InsecureSkipVerify = true
		}
		creds := credentials.NewTLS(tlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	return grpc.Dial(cfg.Server, opts...)
}
