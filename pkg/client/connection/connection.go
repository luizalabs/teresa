package connection

import (
	"github.com/luizalabs/teresa/pkg/client"
	"google.golang.org/grpc"
)

func New(cfgFile, cfgContext string) (*grpc.ClientConn, error) {
	cfg, err := client.GetConfig(cfgFile, cfgContext)
	if err != nil {
		return nil, err
	}
	return client.New(*cfg)
}
