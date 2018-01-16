package deploy

import (
	"fmt"
	"strings"
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
	st "github.com/luizalabs/teresa/pkg/server/storage"
)

func TestNewPodSpec(t *testing.T) {
	expectedAppName := "app-test"
	a := &app.App{
		Name: expectedAppName,
		EnvVars: []*app.EnvVar{
			{Key: "APP-ENV-KEY", Value: "APP-ENV-VALUE"},
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
	opts := &Options{
		SlugBuilderImage: "image",
		BuildLimitCPU:    "800m",
		BuildLimitMemory: "1Gi",
	}

	ps := newBuildSpec(
		&app.App{},
		expectedDeployId,
		expectedTarBallLocation,
		expectedBuildDest,
		st.NewFake(),
		opts,
	)

	if !strings.HasSuffix(ps.Name, expectedDeployId) {
		t.Errorf("expected build-%s, got %s", expectedDeployId, ps.Name)
	}

	if ps.Image != opts.SlugBuilderImage {
		t.Errorf("expected %s, got %s", opts.SlugBuilderImage, ps.Image)
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

	if ps.ContainerLimits.CPU != opts.BuildLimitCPU {
		t.Errorf("expected %s, got %s", opts.BuildLimitCPU, ps.ContainerLimits.CPU)
	}
	if ps.ContainerLimits.Memory != opts.BuildLimitMemory {
		t.Errorf("expected %s, got %s", opts.BuildLimitMemory, ps.ContainerLimits.Memory)
	}
}

func TestNewCommandRunSpec(t *testing.T) {
	expectedSlugURL := "http://teresa.io/slug.tgz"
	expectedCommand := "python manage.py migrate"
	expectedBuildId := "1234"
	a := &app.App{Name: "teresa"}
	s := st.NewFake()
	opts := &Options{
		SlugRunnerImage:  "image",
		BuildLimitCPU:    "800m",
		BuildLimitMemory: "1Gi",
	}

	ps := newRunCommandSpec(a, expectedBuildId, expectedCommand, expectedSlugURL, s, opts)
	if !strings.HasSuffix(ps.Name, fmt.Sprintf("%s-%s", a.Name, expectedBuildId)) {
		t.Errorf("expected release-%s-%s, got %s", a.Name, expectedBuildId, ps.Name)
	}

	if ps.Image != opts.SlugRunnerImage {
		t.Errorf("expected %s, got %s", opts.SlugRunnerImage, ps.Image)
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

	if ps.ContainerLimits.CPU != opts.BuildLimitCPU {
		t.Errorf("expected %s, got %s", opts.BuildLimitCPU, ps.ContainerLimits.CPU)
	}
	if ps.ContainerLimits.Memory != opts.BuildLimitMemory {
		t.Errorf("expected %s, got %s", opts.BuildLimitMemory, ps.ContainerLimits.Memory)
	}
}

func TestNewDeploySpec(t *testing.T) {
	expectedDescription := "test"
	expectedSlugURL := "http://teresa.io/slug.tgz"
	expectedProcessType := "worker"
	expectedName := "deploy-test"
	opts := &Options{RevisionHistoryLimit: 5}

	ds := newDeploySpec(
		&app.App{Name: expectedName},
		&TeresaYaml{},
		st.NewFake(),
		expectedDescription,
		expectedSlugURL,
		expectedProcessType,
		opts,
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

	if ds.RevisionHistoryLimit != opts.RevisionHistoryLimit {
		t.Errorf("expected %d, got %d", opts.RevisionHistoryLimit, ds.RevisionHistoryLimit)
	}
}

func TestNewDeploySpecDefaultDrainTimeout(t *testing.T) {
	expectedDescription := "test"
	expectedSlugURL := "http://teresa.io/slug.tgz"
	expectedProcessType := "worker"
	expectedName := "deploy-test"
	opts := &Options{RevisionHistoryLimit: 5}

	ds := newDeploySpec(
		&app.App{Name: expectedName},
		&TeresaYaml{},
		st.NewFake(),
		expectedDescription,
		expectedSlugURL,
		expectedProcessType,
		opts,
	)

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
