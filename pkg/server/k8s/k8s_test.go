package k8s

import (
	"testing"
)

func TestK8sNew(t *testing.T) {
	var testCases = []struct {
		serviceType   string
		expectedError error
	}{
		{"LoadBalancer", nil},
		{"InvalidType", ErrInvalidServiceType},
	}

	for _, tc := range testCases {
		conf := &Config{DefaultServiceType: tc.serviceType}
		if _, err := New(conf); err != tc.expectedError {
			t.Errorf("expected %v, got %v", tc.expectedError, err)
		}
	}
}
