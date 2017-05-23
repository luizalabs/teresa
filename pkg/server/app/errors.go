package app

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrAppAlreadyExists = status.Errorf(codes.AlreadyExists, "App already exists")
	ErrAppNotFound      = status.Errorf(codes.NotFound, "App not found")
)
