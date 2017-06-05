package app

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrAlreadyExists = status.Errorf(codes.AlreadyExists, "App already exists")
	ErrNotFound      = status.Errorf(codes.NotFound, "App not found")
	ErrInvalidApp    = status.Errorf(codes.InvalidArgument, "Invalid App")
)
