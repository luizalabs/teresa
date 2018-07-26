package spec

import (
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

func TestRunnerPodBuilder(t *testing.T) {
	expectedName := "runner"
	expectedImage := "runner/image"
	expectedInitImage := "init/image"
	expectedSlug := "nowhere"
	expectedLimitCPU := "800m"
	expectedLimitMemory := "1Gi"
	expectedNginxImage := "nginx/image"
	expectedArgs := []string{"start", "web"}
	expectedLabels := map[string]string{"run": "app"}

	a := &app.App{
		Name:        "test",
		ProcessType: app.ProcessTypeWeb,
		EnvVars:     []*app.EnvVar{&app.EnvVar{"ENV", "VAR"}},
		SecretFiles: []string{"secret", "file"},
	}

	ps := NewRunnerPodBuilder(expectedName, expectedImage, expectedInitImage).
		ForApp(a).
		WithSlug(expectedSlug).
		WithLimits(expectedLimitCPU, expectedLimitMemory).
		WithStorage(storage.NewFake()).
		WithLabels(expectedLabels).
		WithArgs(expectedArgs).
		WithNginxSideCar(expectedNginxImage).
		Build()

	if ps.Name != expectedName {
		t.Errorf("expected %s, got %s", expectedName, ps.Name)
	}
	if actual := len(ps.Containers); actual != 2 {
		t.Fatalf("expected 2 container, got %d", actual)
	}
	if actual := ps.Containers[0].Image; actual != expectedImage {
		t.Errorf("expected %s, got %s", expectedImage, actual)
	}
	if actual := ps.Containers[0].ContainerLimits.CPU; actual != expectedLimitCPU {
		t.Errorf("expected %s, got %s", expectedLimitCPU, actual)
	}
	if actual := ps.Containers[0].ContainerLimits.Memory; actual != expectedLimitMemory {
		t.Errorf("expected %s, got %s", expectedLimitMemory, actual)
	}
	for _, ev := range a.EnvVars {
		if actual := ps.Containers[0].Env[ev.Key]; actual != ev.Value {
			t.Errorf("expected %s, got %s", ev.Value, actual)
		}
	}
	for k, v := range expectedLabels {
		if actual := ps.Labels[k]; actual != v {
			t.Errorf("expected %s for %s, got %s", v, k, actual)
		}
	}
	for i, arg := range expectedArgs {
		if actual := ps.Containers[0].Args[i]; actual != arg {
			t.Errorf("expected %s, got %s", arg, actual)
		}
	}
	if actual := len(ps.InitContainers); actual != 1 {
		t.Errorf("expected at least 1 init container, got %d", actual)
	}
	if actual := ps.InitContainers[0].Image; actual != expectedInitImage {
		t.Errorf("expected %s, got %s", expectedInitImage, actual)
	}
	if actual := ps.InitContainers[0].Env["SLUG_URL"]; actual != expectedSlug {
		t.Errorf("expected %s, got %s", expectedSlug, actual)
	}
	if actual := ps.Containers[1].Image; actual != expectedNginxImage {
		t.Errorf("expected %s, got %s", expectedNginxImage, actual)
	}
	if actual := len(ps.Containers[0].VolumeMounts); actual != 3 {
		t.Errorf("expected volumes of init, nginx and secrets, got %d", actual)
	}
	if actual := ps.Containers[0].VolumeMounts[0].Name; actual != AppSecretName {
		t.Errorf("expected %s, got %s", AppSecretName, actual)
	}
	if actual := ps.Containers[0].VolumeMounts[1].Name; actual != slugVolumeName {
		t.Errorf("expected %s, got %s", slugVolumeName, actual)
	}
}
