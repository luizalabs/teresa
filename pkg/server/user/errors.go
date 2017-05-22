package user

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotFound          = status.Errorf(codes.NotFound, "User not found")
	ErrUserAlreadyExists = status.Errorf(codes.AlreadyExists, "User already exists")
	ErrInvalidPassword   = status.Errorf(codes.InvalidArgument, "Invalid password")
)
