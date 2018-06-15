package cmd

import (
	"testing"
)

func TestValidateLimits(t *testing.T) {
	var testCases = []struct {
		cpu    string
		maxCPU string
		mem    string
		maxMem string
		ok     bool
	}{
		{"100m", "200m", "512Mi", "512Mi", true},
		{"300m", "200m", "512Mi", "512Mi", false},
		{"zzzz", "200m", "512Mi", "512Mi", false},
		{"100m", "200m", "1Gi", "512Mi", false},
		{"100m", "200m", "512Mi", "1Gi", true},
	}

	for _, tc := range testCases {
		lims := newLimits(tc.cpu, tc.maxCPU, tc.mem, tc.maxMem)
		err := ValidateLimits(lims)
		if err != nil && tc.ok {
			t.Error("got unexpected error:", err)
		}
		if err == nil && !tc.ok {
			t.Error("got nil; want error")
		}
	}
}
