package k8s

import (
	"testing"
)

func TestK8sNew(t *testing.T) {
	conf := &Config{DefaultServiceType: "LoadBalancer"}

	if _, err := New(conf); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestK8sNewInvalidDefaultServiceType(t *testing.T) {
	conf := &Config{DefaultServiceType: "Invalid type"}

	if _, err := New(conf); err == nil {
		t.Error("expected error, got nil")
	}
}
