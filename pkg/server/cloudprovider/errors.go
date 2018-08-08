package cloudprovider

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotImplemented          = status.Errorf(codes.Unimplemented, "Operation not implemented for this cloud provider")
	ErrServiceNotFound         = status.Errorf(codes.NotFound, "Service not found")
	ErrNotImplementedOnIngress = status.Errorf(codes.Unimplemented, "Operation not implemented for apps using ingress")
)
