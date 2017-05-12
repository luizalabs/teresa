package client

import (
	"errors"
	"testing"

	"github.com/luizalabs/teresa-api/pkg/server/auth"
)

func TestGetErrorMsg(t *testing.T) {
	var testCases = []struct {
		err      error
		expected string
	}{
		{auth.ErrPermissionDenied, "Permission Denied"},
		{errors.New("Generic Error"), "Generic Error"},
	}

	for _, tc := range testCases {
		actual := GetErrorMsg(tc.err)
		if actual != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, actual)
		}
	}

}
