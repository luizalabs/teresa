package app

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrAlreadyExists = status.Errorf(codes.AlreadyExists, "App already exists")
	ErrNotFound      = status.Errorf(codes.NotFound, "App not found")
	ErrUnknown       = status.Errorf(codes.Unknown, "Unknown error")
)

func newAppErr(grpc error, err error) error {
	return &appErr{err: err, grpc: grpc}
}

type appErr struct {
	err  error // low level
	grpc error
}

func (a *appErr) Error() string { return a.grpc.Error() + ": " + a.err.Error() }
func (a *appErr) Grpc() error   { return a.grpc }

func grpcErr(err error) error {
	type grpcer interface {
		Grpc() error
	}
	grpc, ok := err.(grpcer)
	if ok {
		return grpc.Grpc()
	}
	return nil
}
