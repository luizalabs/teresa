package teresa_errors

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGet(t *testing.T) {
	someErr := errors.New("some err")
	grpcErr := status.Errorf(codes.Unknown, "grpc error")

	var testCases = []struct {
		err      error
		expected error
	}{
		{New(grpcErr, someErr), grpcErr},
		{someErr, someErr},
	}

	for _, tc := range testCases {
		actual := Get(tc.err)
		if actual != tc.expected {
			t.Errorf("expected %v, got %v", actual, tc.expected)
		}
	}
}
