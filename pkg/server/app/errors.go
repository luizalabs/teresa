package app

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrAlreadyExists           = status.Errorf(codes.AlreadyExists, "App already exists")
	ErrNotFound                = status.Errorf(codes.NotFound, "App not found")
	ErrProtectedEnvVar         = status.Errorf(codes.InvalidArgument, "Can't change protected env vars")
	ErrInvalidName             = status.Errorf(codes.InvalidArgument, "Invalid App Name")
	ErrInvalidLimits           = status.Errorf(codes.InvalidArgument, "Invalid Limits")
	ErrInvalidAutoscale        = status.Errorf(codes.InvalidArgument, "Invalid Autoscale")
	ErrInvalidEnvVarName       = status.Errorf(codes.InvalidArgument, "Invalid Env Var Name")
	ErrInvalidSecretName       = status.Errorf(codes.InvalidArgument, "Invalid Secret Name")
	ErrInvalidActionForCronJob = status.Errorf(codes.InvalidArgument, "Invalid action for a cronjob app")
	ErrMissingVirtualHost      = status.Errorf(
		codes.InvalidArgument,
		"Missing --vhost argument with the application domain",
	)
)
