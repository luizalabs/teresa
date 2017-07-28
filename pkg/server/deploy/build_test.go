package deploy

import (
	"strings"
	"testing"

	"github.com/luizalabs/teresa-api/pkg/server/app"
	st "github.com/luizalabs/teresa-api/pkg/server/storage"
)

func TestNewPodSpec(t *testing.T) {
	expectedAppName := "app-test"
	a := &app.App{
		Name: expectedAppName,
		EnvVars: []*app.EnvVar{
			&app.EnvVar{Key: "APP-ENV-KEY", Value: "APP-ENV-VALUE"},
		},
	}
	ev := map[string]string{"ENV-KEY": "ENV-VALUE"}
	expectedName := "test"
	expectedImage := "docker/teresa-test:0.0.1"

	ps := newPodSpec(expectedName, expectedImage, a, ev, st.NewFake())
	if ps.Name != expectedName {
		t.Errorf("expected %s, got %s", expectedName, ps.Name)
	}
	if ps.Namespace != a.Name {
		t.Errorf("expected %s, got %s", a.Name, ps.Namespace)
	}
	if ps.Image != expectedImage {
		t.Errorf("expected %s, got %s", expectedImage, ps.Image)
	}

	for _, env := range a.EnvVars {
		if ps.Env[env.Key] != env.Value {
			t.Errorf("expected %s, got %s for key %s", env.Value, ps.Env[env.Key], env.Key)
		}
	}

	for k, v := range ev {
		if ps.Env[k] != v {
			t.Errorf("expected %s, got %s for key %s", v, ps.Env[k], k)
		}
	}
}

func TestNewBuildSpec(t *testing.T) {
	expectedDeployId := "123"
	expectedTarBallLocation := "narnia"
	expectedBuildDest := "nowhere"

	ps := newBuildSpec(
		&app.App{},
		expectedDeployId,
		expectedTarBallLocation,
		expectedBuildDest,
		st.NewFake(),
	)

	if !strings.HasSuffix(ps.Name, expectedDeployId) {
		t.Errorf("expected build-%s, got %s", expectedDeployId, ps.Name)
	}

	if ps.Image != slugBuilderImage {
		t.Errorf("expected %s, got %s", slugBuilderImage, ps.Image)
	}

	ev := map[string]string{
		"TAR_PATH": expectedTarBallLocation,
		"PUT_PATH": expectedBuildDest,
	}
	for k, v := range ev {
		if ps.Env[k] != v {
			t.Errorf("expected %s, got %s for key %s", v, ps.Env[k], k)
		}
	}
}

func TestNewCommandRunSpec(t *testing.T) {
	expectedSlugURL := "http://teresa.io/slug.tgz"
	a := &app.App{Name: "teresa"}
	expectedCommand := "python manage.py migrate"
	s := st.NewFake()

	ps := newRunCommandSpec(a, expectedCommand, expectedSlugURL, s)
	if !strings.HasSuffix(ps.Name, a.Name) {
		t.Errorf("expected release-%s, got %s", a.Name, ps.Name)
	}

	if ps.Image != slugRunnerImage {
		t.Errorf("expected %s, got %s", slugRunnerImage, ps.Image)
	}

	ev := map[string]string{
		"SLUG_URL":        expectedSlugURL,
		"BUILDER_STORAGE": s.Type(),
	}
	for k, v := range ev {
		if ps.Env[k] != v {
			t.Errorf("expected %s, got %s for key %s", v, ps.Env[k], k)
		}
	}
}

func TestNewDeploySpec(t *testing.T) {
	expectedDescription := "test"
	expectedSlugURL := "http://teresa.io/slug.tgz"
	expectedProcessType := "worker"
	expectedName := "deploy-test"
	expectedRevisionHistoryLimit := 5

	ds := newDeploySpec(
		&app.App{Name: expectedName},
		&TeresaYaml{},
		st.NewFake(),
		expectedDescription,
		expectedSlugURL,
		expectedProcessType,
		expectedRevisionHistoryLimit,
	)

	if len(ds.Args) != 2 || ds.Args[1] != expectedProcessType {
		t.Errorf("expected [start %s], got %v", expectedProcessType, ds.Args)
	}

	if ds.SlugURL != expectedSlugURL {
		t.Errorf("expected %s, got %s", expectedSlugURL, ds.SlugURL)
	}

	if ds.Description != expectedDescription {
		t.Errorf("expected %s, got %s", expectedDescription, ds.Description)
	}

	if ds.PodSpec.Name != expectedName {
		t.Errorf("expected %s, got %s", expectedName, ds.PodSpec.Name)
	}

	if ds.RevisionHistoryLimit != expectedRevisionHistoryLimit {
		t.Errorf("expected %d, got %d", expectedRevisionHistoryLimit, ds.RevisionHistoryLimit)
	}
}
