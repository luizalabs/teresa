package k8s

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8serrors "k8s.io/client-go/pkg/api/errors"
)

var (
	ErrInvalidServiceType = errors.New("Invalid service type")
	ErrNotFound           = status.Errorf(codes.NotFound, "Resource not found")
)

func cleanError(err error) error {
	if e, ok := err.(*k8serrors.StatusError); ok {
		switch e.ErrStatus.Code {
		case 404:
			return ErrNotFound
		default:
			return e
		}
	}
	return err
}
