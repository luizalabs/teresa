package validation

import "testing"

func TestIsEnvVarName(t *testing.T) {
	var testCases = []struct {
		name string
		res  bool
	}{
		{"TEST", true},
		{"TEST_TERESA", true},
		{"@luizalabs", false},
		{"TEST:TERESA", false},
		{"TEST/TERESA", false},
	}

	for _, tc := range testCases {
		if b := IsEnvVarName(tc.name); b != tc.res {
			t.Errorf("want %v; got %v (name: %s)", tc.res, b, tc.name)
		}
	}
}

func TestIsProtectedEnvVar(t *testing.T) {
	var testCases = []struct {
		name string
		res  bool
	}{
		{"TEST", false},
		{"PYTHONPATH", true},
	}

	for _, tc := range testCases {
		if b := IsProtectedEnvVar(tc.name); b != tc.res {
			t.Errorf("want %v; got %v (name: %s)", tc.res, b, tc.name)
		}
	}
}
