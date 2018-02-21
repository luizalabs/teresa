package spec

import (
	"strings"
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

func TestNewCronJobSpec(t *testing.T) {
	expectedImage := "image"
	expectedDescription := "test"
	expectedSlugURL := "http://teresa.io/slug.tgz"
	a := &app.App{Name: "cron-test"}
	expectedCommand := "happy cron job"
	expectedSchedule := "*/1 * * * *"

	cs := NewCronJob(
		expectedImage,
		expectedDescription,
		expectedSlugURL,
		expectedSchedule,
		a,
		storage.NewFake(),
		strings.Split(expectedCommand, " ")...,
	)

	if len(cs.Args) != 3 || strings.Join(cs.Args, " ") != expectedCommand {
		t.Errorf("expected [%s], got %v", expectedCommand, cs.Args)
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
}
