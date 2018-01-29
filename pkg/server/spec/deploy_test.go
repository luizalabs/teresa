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

	ds := NewDeploy(
		expectedImage,
		expectedDescription,
		expectedSlugURL,
		expectedRevisionHistoryLimit,
		a,
		&TeresaYaml{},
		storage.NewFake(),
	)

	if len(ds.Args) != 2 || ds.Args[1] != a.ProcessType {
		t.Errorf("expected [start %s], got %v", a.ProcessType, ds.Args)
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
