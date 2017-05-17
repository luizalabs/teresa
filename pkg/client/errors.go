package client

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

func GetErrorMsg(err error) string {
	switch grpc.Code(err) {
	case codes.PermissionDenied:
		return "Permission Denied"
	case codes.Unavailable:
		return "Server Unavailable"
	case codes.AlreadyExists:
		return "Resource Already Exists"
	}
	return err.Error()
}
