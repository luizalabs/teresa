package spec

import (
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

func TestDeployBuilder(t *testing.T) {
	expectedPodName := "test"
	expectedNamespace := "ns"
	pod := NewRunnerPodBuilder(expectedPodName, "runner", "store").
		ForApp(&app.App{Name: expectedNamespace}).
		WithStorage(storage.NewFake()).
		Build()

	expectedSlugURL := "some/slug.tgz"
	expectedDescription := "teste"
	expectedRevisionHistoryLimit := 5
	expectedMatchLabels := map[string]string{"expected": "label"}
	ds := NewDeployBuilder(expectedSlugURL).
		WithPod(pod).
		WithDescription(expectedDescription).
		WithRevisionHistoryLimit(expectedRevisionHistoryLimit).
		WithTeresaYaml(&TeresaYaml{}).
		WithMatchLabels(expectedMatchLabels).
		Build()

	if ds.Pod.Name != expectedPodName {
		t.Errorf("expected %s, got %s", expectedPodName, ds.Pod.Name)
	}
	if ds.Pod.Namespace != expectedNamespace {
		t.Errorf("expected %s, got %s", expectedNamespace, ds.Pod.Namespace)
	}
	if ds.SlugURL != expectedSlugURL {
		t.Errorf("expected %s, got %s", expectedSlugURL, ds.SlugURL)
	}
	if ds.Description != expectedDescription {
		t.Errorf("expected %s, got %s", expectedDescription, ds.Description)
	}
	if ds.RevisionHistoryLimit != expectedRevisionHistoryLimit {
		t.Errorf("expected %d, got %d", expectedRevisionHistoryLimit, ds.RevisionHistoryLimit)
	}
	for k, v := range expectedMatchLabels {
		if actual := ds.MatchLabels[k]; actual != v {
			t.Errorf("expected %s for key %s, got %s", v, k, actual)
		}
	}
	if ds.Lifecycle == nil {
		t.Fatal("expected lifecycle; got nil")
	}

	if ds.Lifecycle.PreStop == nil {
		t.Fatal("expected prestop; got nil")
	}

	if ds.Lifecycle.PreStop.DrainTimeoutSeconds != defaultDrainTimeoutSeconds {
		t.Errorf("got %d; want %d", ds.Lifecycle.PreStop.DrainTimeoutSeconds, defaultDrainTimeoutSeconds)
	}
}
