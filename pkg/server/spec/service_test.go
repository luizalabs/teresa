package spec

import "testing"

func TestNewDefaultService(t *testing.T) {
	expectedAppName := "test"
	expectedServiceType := "LoadBalancer"

	ss := NewDefaultService(expectedAppName, expectedServiceType)
	if ss.Namespace != expectedAppName {
		t.Errorf("expected %s, got %s", expectedAppName, ss.Namespace)
	}
	if ss.Name != expectedAppName {
		t.Errorf("expected %s, got %s", expectedAppName, ss.Name)
	}
	if ss.Type != expectedServiceType {
		t.Errorf("expected %s, got %s", expectedServiceType, ss.Type)
	}
	if ss.Ports[0].TargetPort != DefaultPort {
		t.Errorf("expected %d, got %d", DefaultPort, ss.Ports[0].TargetPort)
	}
	if name := ss.Labels["run"]; name != expectedAppName {
		t.Errorf("expected label run with value %s, got %s", expectedAppName, name)
	}
}
