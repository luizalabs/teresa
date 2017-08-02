package k8s

import (
	"testing"
)

func TestValidateConfig(t *testing.T) {
	var testCases = []struct {
		serviceType   string
		expectedError error
	}{
		{"LoadBalancer", nil},
		{"InvalidType", ErrInvalidServiceType},
	}

	for _, tc := range testCases {
		conf := &Config{ConfigFile: "test", DefaultServiceType: tc.serviceType}
		if err := validateConfig(conf); err != tc.expectedError {
			t.Errorf("expected %v, got %v", tc.expectedError, err)
		}
	}
}