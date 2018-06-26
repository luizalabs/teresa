package service

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotFound               = status.Errorf(codes.NotFound, "Service not found")
	ErrInvalidSourceRanges    = status.Errorf(codes.InvalidArgument, "Invalid source ranges")
	ErrWhitelistUnimplemented = status.Errorf(codes.Unimplemented, "Whitelist not supported for ingress or internal apps")
)
