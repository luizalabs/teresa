package exec

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrDeployNotFound  = status.Errorf(codes.NotFound, "Current deploy not found")
	ErrNonZeroExitCode = status.Errorf(codes.Unknown, "Exec command returned a non zero value")
	ErrTimeout         = status.Errorf(codes.Aborted, "Timeout on performing the command")
)
