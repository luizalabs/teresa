package storage

import (
	"testing"
)

func TestNewStorage(t *testing.T) {
	conf := &Config{Type: "s3"}

	if _, err := New(conf); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestNewStorageInvalidType(t *testing.T) {
	conf := &Config{Type: "Invalid type"}

	if _, err := New(conf); err == nil {
		t.Error("expected error, got nil")
	}
}
