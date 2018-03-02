package spec

import (
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

const (
	slugVolumeName      = "slug"
	slugVolumeMountPath = "/slug"
)

type Volume struct {
	Name       string
	SecretName string
	EmptyDir   bool
}

type Pod struct {
	Name           string
	Namespace      string
	Containers     []*Container
	Volumes        []*Volume
	InitContainers []*Container
}

func NewPod(name, image string, a *app.App, envVars map[string]string, fs storage.Storage) *Pod {
	ps := &Pod{
		Name:      name,
		Namespace: a.Name,
		Containers: []*Container{{
			Name:  name,
			Image: image,
			Env:   envVars,
		}},
		Volumes: []*Volume{
			{
				Name:       "storage-keys",
				SecretName: fs.K8sSecretName(),
			},
			{
				Name:     slugVolumeName,
				EmptyDir: true,
			},
		},
	}
	for _, e := range a.EnvVars {
		ps.Containers[0].Env[e.Key] = e.Value
	}
	for k, v := range fs.PodEnvVars() {
		ps.Containers[0].Env[k] = v
	}
	return ps
}

func NewBuilder(name, tarBallLocation, buildDest, image string, a *app.App, fs storage.Storage, cl *ContainerLimits) *Pod {
	ps := NewPod(
		name,
		image,
		a,
		map[string]string{
			"TAR_PATH":        tarBallLocation,
			"PUT_PATH":        buildDest,
			"BUILDER_STORAGE": fs.Type(),
		},
		fs,
	)
	ps.Containers[0].VolumeMounts = []*VolumeMounts{newStorageKeyVolumeMount()}
	ps.Containers[0].ContainerLimits = cl
	return ps
}

func NewRunner(name, slugURL string, imgs *SlugImages, a *app.App, fs storage.Storage, cl *ContainerLimits, command ...string) *Pod {
	ps := NewPod(
		name,
		imgs.Runner,
		a,
		map[string]string{
			"APP":      a.Name,
			"SLUG_URL": slugURL,
			"SLUG_DIR": slugVolumeMountPath,
		},
		fs,
	)
	ps.Containers[0].Args = command
	ps.Containers[0].ContainerLimits = cl
	ps.Containers[0].VolumeMounts = []*VolumeMounts{newSlugVolumeMount()}
	ps.InitContainers = newInitContainers(slugURL, imgs.Store, a, fs)
	return ps
}
