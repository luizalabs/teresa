package client

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

func TestGetErrorMsg(t *testing.T) {
	var testCases = []struct {
		err      error
		expected string
	}{
		{auth.ErrPermissionDenied, "Permission Denied"},
		{status.Errorf(codes.Unavailable, "Server Unavailable"), "Server Unavailable"},
		{errors.New("Generic Error"), "Generic Error"},
	}

	for _, tc := range testCases {
		actual := GetErrorMsg(tc.err)
		if actual != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, actual)
		}
	}

}
