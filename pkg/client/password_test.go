package client

import (
	"testing"
)

func TestCheckPassword(t *testing.T) {
	var testCases = []struct {
		pass        string
		expectError bool
	}{
		{"12345678", false},
		{"test", true},
	}

	for _, tc := range testCases {
		if err := checkPassword(tc.pass); (err != nil && !tc.expectError) || (err == nil && tc.expectError) {
			t.Errorf("expect error %v, got error %v", tc.expectError, err)
		}
	}
}
