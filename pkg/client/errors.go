package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func GetErrorMsg(err error) string {
	switch grpc.Code(err) {
	case codes.PermissionDenied:
		return "Permission Denied"
	}
	return err.Error()
}
