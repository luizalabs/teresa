package connection

import (
	"github.com/luizalabs/teresa-api/pkg/client"
	"google.golang.org/grpc"
)

type Options struct {
	NoTLS       bool
	TLSInsecure bool
}

func New(cfgFile string, opts *Options) (*grpc.ClientConn, error) {
	cfg, err := client.GetConfig(cfgFile)
	if err != nil {
		return nil, err
	}
	return client.New(*cfg, opts.NoTLS, opts.TLSInsecure)
}
