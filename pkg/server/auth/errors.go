package auth

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrPermissionDenied = status.Errorf(codes.PermissionDenied, "Permission Denied")
