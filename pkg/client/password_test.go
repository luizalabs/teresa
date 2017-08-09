package client

import (
	"testing"
)

func TestCheckPassword(t *testing.T) {
	var testCases = []struct {
		pass string
		ok   bool
	}{
		{"12345678", true},
		{"test", false},
	}

	for _, tc := range testCases {
		if err := checkPassword(tc.pass); err != nil && tc.ok {
			t.Errorf("expected failure, got success")
		}
	}
}
