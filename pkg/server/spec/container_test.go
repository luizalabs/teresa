package spec

import "testing"

func TestContainerBuilder(t *testing.T) {
	expectedName := "test"
	expectedImage := "test/image:v1"
	expectedEnv := map[string]string{"Key": "Value"}
	expectedSecrets := []string{"SECRET", "VERY-SECRET"}
	expectedLimits := ContainerLimits{CPU: "200m", Memory: "256Mi"}
	expectedCmd := []string{"echo", "hi"}
	expectedArgs := []string{"hello", "from", "test"}
	expectedPort := 5000

	c := NewContainerBuilder(expectedName, expectedImage).
		WithEnv(expectedEnv).
		WithSecrets(expectedSecrets).
		WithLimits(expectedLimits.CPU, expectedLimits.Memory).
		WithCommand(expectedCmd).
		WithArgs(expectedArgs).
		ExposePort("http", expectedPort).
		Build()

	if actual := c.Name; actual != expectedName {
		t.Errorf("expected %s, got %s", expectedName, actual)
	}
	if actual := c.Image; actual != expectedImage {
		t.Errorf("expected %s, got %s", expectedImage, actual)
	}
	if actual := c.ContainerLimits.CPU; actual != expectedLimits.CPU {
		t.Errorf("expected %s, got %s", expectedLimits.CPU, actual)
	}
	if actual := c.ContainerLimits.Memory; actual != expectedLimits.Memory {
		t.Errorf("expected %s, got %s", expectedLimits.Memory, actual)
	}
	if actual := c.Ports[0].ContainerPort; actual != int32(expectedPort) {
		t.Errorf("expected %d, got %d", expectedPort, actual)
	}
	for k, v := range expectedEnv {
		if actual := c.Env[k]; actual != v {
			t.Errorf("expected value %s for key %s, got %s", v, k, actual)
		}
	}
	for i, s := range expectedSecrets {
		if actual := c.Secrets[i]; actual != s {
			t.Errorf("expected %s, got %s", s, actual)
		}
	}
	for i, cmd := range expectedCmd {
		if actual := c.Command[i]; actual != cmd {
			t.Errorf("expected %s, got %s", cmd, actual)
		}
	}
	for i, arg := range expectedArgs {
		if actual := c.Args[i]; actual != arg {
			t.Errorf("expected %s, got %s", arg, actual)
		}
	}
}
