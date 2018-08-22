package k8s

import (
	"github.com/pkg/errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

var (
	ErrInvalidServiceType = errors.New("Invalid service type")
	ErrNotFound           = status.Errorf(codes.NotFound, "Resource not found")
	ErrPodRunFailed       = status.Errorf(codes.Aborted, "Pod went into failed status")
	ErrPodStillRunning    = status.Errorf(codes.Unknown, "Pod still running")
)

func (k *Client) IsNotFound(err error) bool {
	return k8serrors.IsNotFound(errors.Cause(err))
}

func (k *Client) IsAlreadyExists(err error) bool {
	return k8serrors.IsAlreadyExists(errors.Cause(err))
}

func (k *Client) IsInvalid(err error) bool {
	return k8serrors.IsInvalid(errors.Cause(err))
}

func (k *Client) IsUnknown(err error) bool {
	_, ok := errors.Cause(err).(k8serrors.APIStatus)
	return !ok
}
