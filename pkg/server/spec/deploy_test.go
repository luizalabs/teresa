package spec

import (
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

func TestNewDeploySpec(t *testing.T) {
	expectedImage := "image"
	expectedDescription := "test"
	expectedSlugURL := "http://teresa.io/slug.tgz"
	expectedRevisionHistoryLimit := 5
	a := &app.App{Name: "deploy-test", ProcessType: "worker"}
	imgs := &Images{SlugRunner: expectedImage}

	ds := NewDeploy(
		imgs,
		"",
		expectedDescription,
		expectedSlugURL,
		expectedRevisionHistoryLimit,
		a,
		&TeresaYaml{},
		storage.NewFake(),
	)

	if len(ds.Containers[0].Args) != 2 || ds.Containers[0].Args[1] != a.ProcessType {
		t.Errorf("expected [start %s], got %v", a.ProcessType, ds.Containers[0].Args)
	}

	if ds.SlugURL != expectedSlugURL {
		t.Errorf("expected %s, got %s", expectedSlugURL, ds.SlugURL)
	}

	if ds.Description != expectedDescription {
		t.Errorf("expected %s, got %s", expectedDescription, ds.Description)
	}

	if ds.Pod.Name != a.Name {
		t.Errorf("expected %s, got %s", a.Name, ds.Pod.Name)
	}

	if ds.RevisionHistoryLimit != expectedRevisionHistoryLimit {
		t.Errorf("expected %d, got %d", expectedRevisionHistoryLimit, ds.RevisionHistoryLimit)
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

func TestNewDeploySpecInitContainers(t *testing.T) {
	expectedImage := "image"
	a := &app.App{}
	imgs := &Images{SlugStore: expectedImage}

	ds := NewDeploy(imgs, "", "", "", 0, a, &TeresaYaml{}, storage.NewFake())

	if len(ds.InitContainers) != 1 {
		t.Errorf("got %d; want %d", len(ds.InitContainers), 1)
	}
	if ds.InitContainers[0].Image != expectedImage {
		t.Errorf("got %s; want %s", ds.InitContainers[0].Image, expectedImage)
	}
}
