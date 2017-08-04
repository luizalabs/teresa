package app

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrAlreadyExists    = status.Errorf(codes.AlreadyExists, "App already exists")
	ErrNotFound         = status.Errorf(codes.NotFound, "App not found")
	ErrProtectedEnvVar  = status.Errorf(codes.InvalidArgument, "Can't change protected env vars")
	ErrInvalidLimits    = status.Errorf(codes.InvalidArgument, "Invalid Limits")
	ErrInvalidAutoScale = status.Errorf(codes.InvalidArgument, "Invalid AutoScale")
)
