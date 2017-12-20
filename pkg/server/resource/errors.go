package resource

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrAlreadyExists = status.Errorf(codes.AlreadyExists, "Resource already exists")
	ErrNotFound      = status.Errorf(codes.NotFound, "Resource not found")
)
