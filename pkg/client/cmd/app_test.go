package cmd

import "testing"

var flagTests = []struct {
	min    int32
	max    int32
	output bool
}{
	{5, 10, true},
	{10, 5, false},
	{5, 5, true},
	{5, -10, false},
	{-10, 5, false},
	{-10, -10, false},
}

func TestValidateFlags(t *testing.T) {
	for _, tt := range flagTests {
		_, isValid := validateFlags(tt.min, tt.max)
		if isValid != tt.output {
			t.Errorf("Validate fail, expect %t and get %t (min = %d, max = %d).", tt.output, isValid, tt.min, tt.max)
		}
	}
}
