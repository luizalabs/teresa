package validation

import "testing"

func TestIsValidEmail(t *testing.T) {
	var testCases = []struct {
		email          string
		expectedResult bool
	}{
		{"gopher@luizalabs.com", true},
		{"gopher@magazineluiza.com.br", true},
		{"gopher123@luizalabs.com", true},
		{"gopher.vimmer@luizalabs.com", true},
		{"gopher_vimmer@luizalabs.com", true},
		{"gopher-vimmer@luizalabs.com", true},
		{"gopher", false},
		{"gopher.com", false},
		{"gopher@@luizalabs.com", false},
		{"gopher@.com", false},
	}

	for _, tc := range testCases {
		if b := IsValidEmail(tc.email); b != tc.expectedResult {
			t.Errorf("expected %v; got %v (email: %s)", tc.expectedResult, b, tc.email)
		}
	}
}
