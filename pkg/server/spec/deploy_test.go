package spec

import (
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"

	yaml "gopkg.in/yaml.v2"
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
	expectedDNSConfigNdots := "2"
	ds := NewDeployBuilder(expectedSlugURL).
		WithPod(pod).
		WithDescription(expectedDescription).
		WithRevisionHistoryLimit(expectedRevisionHistoryLimit).
		WithTeresaYaml(&TeresaYaml{}).
		WithDNSConfigNdots(expectedDNSConfigNdots).
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

	if ds.DNSConfig == nil {
		t.Error("expected dnsConfig; got nil")
	}

	if ds.DNSConfig.Options[0].Value != "2" {
		t.Errorf("expected dnsConfig.Options[0].value to be 2; got %s", ds.DNSConfig.Options[0].Value)
	}
}

func TestTeresaYamlOvewriteNdots(t *testing.T) {
	expectedDNSConfigNdots := "2"
	expectedSlugURL := "some/slug.tgz"
	teresaYaml := &TeresaYaml{}
	teresaYaml.DNSConfig = &DNSConfig{
		Options: append([]DNSOptions{}, DNSOptions{
			Name:  "ndots",
			Value: "1",
		}),
	}

	ds := NewDeployBuilder(expectedSlugURL).
		WithTeresaYaml(teresaYaml).
		WithDNSConfigNdots(expectedDNSConfigNdots).
		Build()

	if ds.DNSConfig.Options[0].Value != "1" {
		t.Errorf("expected dnsConfig.Options[0].value to be 1; got %s", ds.DNSConfig.Options[0].Value)
	}
}

func TestRawData(t *testing.T) {
	b := []byte("field1: value1\nfield2: value2")
	raw := new(RawData)
	type T struct {
		Field1 string
		Field2 string
	}
	v := new(T)

	if err := yaml.Unmarshal(b, raw); err != nil {
		t.Fatal("got unexpected error:", err)
	}
	if err := raw.Unmarshal(v); err != nil {
		t.Fatal("got unexpected error:", err)
	}

	if v.Field1 != "value1" {
		t.Errorf("got %s; want %s", v.Field1, "value1")
	}
	if v.Field2 != "value2" {
		t.Errorf("got %s; want %s", v.Field2, "value2")
	}
}
