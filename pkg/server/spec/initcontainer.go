package spec

import (
	"github.com/luizalabs/teresa/pkg/server/storage"
)

const (
	slugVolumeName      = "slug"
	slugVolumeMountPath = "/slug"
)

func NewInitContainer(image, slugURL string, fs storage.Storage) *Container {
	env := map[string]string{
		"BUILDER_STORAGE": fs.Type(),
		"SLUG_URL":        slugURL,
		"SLUG_DIR":        slugVolumeMountPath,
	}
	return NewContainerBuilder("slugstore", image).
		WithEnv(env).
		WithEnv(fs.PodEnvVars()).
		Build()
}
