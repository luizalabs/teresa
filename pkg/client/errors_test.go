package client

import (
	"errors"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

func TestGetErrorMsg(t *testing.T) {
	var testCases = []struct {
		err      error
		expected string
	}{
		{auth.ErrPermissionDenied, "Permission Denied"},
		{grpc.Errorf(codes.Unavailable, ""), "Server Unavailable"},
		{errors.New("Generic Error"), "Generic Error"},
	}

	for _, tc := range testCases {
		actual := GetErrorMsg(tc.err)
		if actual != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, actual)
		}
	}

}
