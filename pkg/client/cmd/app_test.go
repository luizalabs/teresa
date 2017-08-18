package cmd

import "testing"

func TestValidateFlags(t *testing.T) {

	var testCases = []struct {
		min      int32
		max      int32
		expected bool
	}{
		{5, 10, true},
		{10, 5, false},
		{5, 5, true},
		{5, -10, false},
		{-10, 5, false},
		{-10, -10, false},
		{0, 0, false},
		{0, 1, false},
		{1, 0, false},
	}

	for _, tc := range testCases {
		_, isValid := validateFlags(tc.min, tc.max)
		if isValid != tc.expected {
			t.Errorf("expected %t and got %t (min = %d, max = %d).", tc.expected, isValid, tc.min, tc.max)
		}
	}
}
