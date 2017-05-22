package client

import (
	"google.golang.org/grpc/status"
)

func GetErrorMsg(err error) string {
	stat, ok := status.FromError(err)
	if !ok {
		return err.Error()
	}
	return stat.Message()
}
