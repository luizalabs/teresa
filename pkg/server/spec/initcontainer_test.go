package spec

import (
	"testing"

	"github.com/luizalabs/teresa/pkg/server/storage"
)

func TestInitContainerBuilder(t *testing.T) {
	fs := storage.NewFake()
	expectedImage := "slugstore"
	expectedSlugURL := "slug"

	c := NewInitContainer(expectedImage, expectedSlugURL, fs)
	if actual := c.Image; actual != expectedImage {
		t.Errorf("expected %s, got %s", expectedImage, actual)
	}
	if actual := c.Env["SLUG_DIR"]; actual != slugVolumeMountPath {
		t.Errorf("expected %s, got %s", slugVolumeMountPath, actual)
	}
	if actual := c.Env["SLUG_URL"]; actual != expectedSlugURL {
		t.Errorf("expected %s, got %s", expectedSlugURL, actual)
	}
	if actual := c.Env["BUILDER_STORAGE"]; actual != fs.Type() {
		t.Errorf("expected %s, got %s", fs.Type(), actual)
	}
}
