package cloudprovider

import (
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrNotImplemented       = status.Errorf(codes.Unimplemented, "Operation not implemented for this cloud provider")
	ErrInvalidCloudProvider = errors.New("invalid cloud provider")
)
