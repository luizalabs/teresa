package spec

import (
	"strings"
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

func TestNewCronJobSpec(t *testing.T) {
	expectedImage := "image"
	expectedInitImage := "init-image"
	expectedDescription := "test"
	expectedSlugURL := "http://teresa.io/slug.tgz"
	a := &app.App{Name: "cron-test"}
	expectedCommand := "happy cron job"
	expectedSchedule := "*/1 * * * *"

	imgs := &Images{SlugRunner: expectedImage, SlugStore: expectedInitImage}

	cs := NewCronJob(
		expectedDescription,
		expectedSlugURL,
		expectedSchedule,
		imgs,
		a,
		storage.NewFake(),
		strings.Split(expectedCommand, " ")...,
	)

	csArgs := cs.Containers[0].Args
	if len(csArgs) != 3 || strings.Join(csArgs, " ") != expectedCommand {
		t.Errorf("expected [%s], got %v", expectedCommand, csArgs)
	}

	if cs.SlugURL != expectedSlugURL {
		t.Errorf("expected %s, got %s", expectedSlugURL, cs.SlugURL)
	}

	if cs.Description != expectedDescription {
		t.Errorf("expected %s, got %s", expectedDescription, cs.Description)
	}

	if cs.Pod.Name != a.Name {
		t.Errorf("expected %s, got %s", a.Name, cs.Pod.Name)
	}

	if cs.Schedule != expectedSchedule {
		t.Errorf("expected %s, got %s", expectedSchedule, cs.Schedule)
	}

	if len(cs.InitContainers) != 1 {
		t.Errorf("expected %d, got %d", 1, len(cs.InitContainers))
	}
	if cs.InitContainers[0].Image != expectedInitImage {
		t.Errorf("expected %s, got %s", expectedImage, cs.InitContainers[0].Image)
	}
}
