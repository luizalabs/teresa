package spec

import (
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

func newSlugVolumeMount() *VolumeMounts {
	return &VolumeMounts{
		Name:      slugVolumeName,
		MountPath: slugVolumeMountPath,
	}
}

func newStorageKeyVolumeMount() *VolumeMounts {
	return &VolumeMounts{
		Name:      "storage-keys",
		MountPath: "/var/run/secrets/deis/objectstore/creds",
		ReadOnly:  true,
	}
}

func newInitContainers(slugURL, image string, a *app.App, fs storage.Storage) []*Container {
	return []*Container{{
		Name:      "slugstore",
		Namespace: a.Name,
		Image:     image,
		Env: map[string]string{
			"BUILDER_STORAGE": fs.Type(),
			"SLUG_URL":        slugURL,
			"SLUG_DIR":        slugVolumeMountPath,
		},
		VolumeMounts: []*VolumeMounts{
			newStorageKeyVolumeMount(),
			newSlugVolumeMount(),
		},
	}}
}
