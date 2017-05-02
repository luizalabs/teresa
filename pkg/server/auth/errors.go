package auth

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var ErrPermissionDenied = grpc.Errorf(codes.PermissionDenied, "Permission Denied")
