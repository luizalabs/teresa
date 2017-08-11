package client

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/luizalabs/teresa/pkg/server/auth"
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

func TestPrintErrorAndExit(t *testing.T) {
	if os.Getenv("PRINT_ERROR_AND_EXIT") == "1" {
		PrintErrorAndExit("Some terrible error")
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestPrintErrorAndExit")
	cmd.Env = append(os.Environ(), "PRINT_ERROR_AND_EXIT=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Errorf("expected exit status 1, got err %v", err)
}
