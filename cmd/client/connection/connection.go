package connection

import (
	"github.com/luizalabs/teresa-api/pkg/client"
	"google.golang.org/grpc"
)

func New(cfgFile string) (*grpc.ClientConn, error) {
	cfg, err := client.GetConfig(cfgFile)
	if err != nil {
		return nil, err
	}
	return client.New(*cfg)
}
