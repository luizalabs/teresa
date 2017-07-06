package deploy

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrBuildFail = status.Errorf(codes.Unknown, "Build return a non zero value")
