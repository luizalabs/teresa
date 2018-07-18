package spec

import (
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

func TestBuildPodBuider(t *testing.T) {
	expectedName := "builder"
	expectedImage := "builder/image"
	expectedTarBallPath := "narnia"
	expectedBuildDest := "nowhere"
	expectedLimitCPU := "800m"
	expectedLimitMemory := "1Gi"
	a := &app.App{Name: "test", EnvVars: []*app.EnvVar{&app.EnvVar{Key: "k", Value: "v"}}}

	ps := NewBuildPodBuilder(expectedName, expectedImage).
		ForApp(a).
		WithTarBallPath(expectedTarBallPath).
		SendSlugTo(expectedBuildDest).
		WithLimits(expectedLimitCPU, expectedLimitMemory).
		WithStorage(storage.NewFake()).
		Build()

	if ps.Name != expectedName {
		t.Errorf("expected %s, got %s", expectedName, ps.Name)
	}
	if actual := len(ps.Containers); actual != 1 {
		t.Fatalf("expected at least 1 container, got %d", actual)
	}
	if actual := ps.Containers[0].Image; actual != expectedImage {
		t.Errorf("expected %s, got %s", expectedImage, ps.Containers[0].Image)
	}

	ev := map[string]string{
		"TAR_PATH": expectedTarBallPath,
		"PUT_PATH": expectedBuildDest,
	}
	for k, v := range ev {
		if actual := ps.Containers[0].Env[k]; actual != v {
			t.Errorf("expected %s, got %s for key %s", v, actual, k)
		}
	}
	for _, ev := range a.EnvVars {
		if actual := ps.Containers[0].Env[ev.Key]; actual != ev.Value {
			t.Errorf("expected %s for key %s, got %s", ev.Value, ev.Key, actual)
		}
	}
	if actual := ps.Containers[0].ContainerLimits.CPU; actual != expectedLimitCPU {
		t.Errorf("expected %s, got %s", expectedLimitCPU, ps.Containers[0].ContainerLimits.CPU)
	}
	if actual := ps.Containers[0].ContainerLimits.Memory; actual != expectedLimitMemory {
		t.Errorf("expected %s, got %s", expectedLimitMemory, actual)
	}

	if actual := len(ps.Containers[0].VolumeMounts); actual != 1 {
		t.Fatalf("expected at least 1 Volume, got %d", actual)
	}
	if actual := ps.Containers[0].VolumeMounts[0].Name; actual != "storage-keys" {
		t.Errorf("expected %s, got %s", "storage-keys", actual)
	}
}
