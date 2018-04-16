package build

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrBuildFail        = status.Errorf(codes.Unknown, "Build returned a non zero value")
	ErrInvalidBuildName = status.Errorf(codes.InvalidArgument, "Invalid Build Name")
)
