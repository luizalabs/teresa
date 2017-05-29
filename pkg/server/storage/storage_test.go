package storage

import (
	"testing"
)

func TestNewStorage(t *testing.T) {
	var testCases = []struct {
		storageType   storageType
		expectedError error
	}{
		{"s3", nil},
		{"InvalidType", ErrInvalidStorageType},
	}

	for _, tc := range testCases {
		conf := &Config{Type: tc.storageType}
		if _, err := New(conf); err != tc.expectedError {
			t.Errorf("expected %v, got %v", tc.expectedError, err)
		}
	}
}
