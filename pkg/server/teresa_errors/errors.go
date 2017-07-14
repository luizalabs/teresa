package teresa_errors

import "fmt"

type GrpcError interface {
	Grpc() error
}

type Error interface {
	GrpcError
	Error() string
}

type serverError struct {
	err  error
	grpc error
}

func (s *serverError) Error() string { return fmt.Sprintf("%s: %s", s.grpc.Error(), s.err.Error()) }
func (s *serverError) Grpc() error   { return s.grpc }

// Get returns a gRPC error if err is a GrpcError, otherwise, returns the low level error
func Get(err error) error {
	if grpc, ok := err.(GrpcError); ok {
		return grpc.Grpc()
	}
	return err
}

func New(grpcErr, err error) Error {
	return &serverError{grpc: grpcErr, err: err}
}
