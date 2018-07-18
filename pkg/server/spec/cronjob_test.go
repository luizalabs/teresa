package spec

import (
	"strings"
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

func TestNewCronJobSpec(t *testing.T) {
	expectedPodName := "test"
	expectedNamespace := "ns"
	expectedCommand := "happy cron job"
	pod := NewRunnerPodBuilder(expectedPodName, "runner", "store").
		ForApp(&app.App{Name: expectedNamespace}).
		WithStorage(storage.NewFake()).
		WithArgs(strings.Split(expectedCommand, " ")).
		Build()

	expectedDescription := "test"
	expectedSchedule := "*/1 * * * *"
	expectedSlugURL := "some/slug.tgz"
	cs := NewCronJobBuilder(expectedSlugURL).
		WithPod(pod).
		WithDescription(expectedDescription).
		WithSchedule(expectedSchedule).
		Build()

	if cs.Pod.Name != expectedPodName {
		t.Errorf("expected %s, got %s", expectedPodName, cs.Pod.Name)
	}
	if cs.Pod.Namespace != expectedNamespace {
		t.Errorf("expected %s, got %s", expectedNamespace, cs.Pod.Namespace)
	}
	if cs.SlugURL != expectedSlugURL {
		t.Errorf("expected %s, got %s", expectedSlugURL, cs.SlugURL)
	}
	if cs.Description != expectedDescription {
		t.Errorf("expected %s, got %s", expectedDescription, cs.Description)
	}

	csArgs := cs.Containers[0].Args
	if len(csArgs) != 3 || strings.Join(csArgs, " ") != expectedCommand {
		t.Errorf("expected [%s], got %v", expectedCommand, csArgs)
	}

	if cs.Schedule != expectedSchedule {
		t.Errorf("expected %s, got %s", expectedSchedule, cs.Schedule)
	}

}
