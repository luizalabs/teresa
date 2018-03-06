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
	Name          string
	SecretName    string
	ConfigMapName string
	EmptyDir      bool
}

type Pod struct {
	Name           string
	Namespace      string
	Containers     []*Container
	Volumes        []*Volume
	InitContainers []*Container
}

func newPodVolumes(appName string, fs storage.Storage, hasNginx bool) []*Volume {
	volumes := []*Volume{
		{
			Name:       "storage-keys",
			SecretName: fs.K8sSecretName(),
		},
		{
			Name:     slugVolumeName,
			EmptyDir: true,
		},
	}
	if hasNginx {
		volumes = append(volumes, &Volume{
			Name:          "nginx-conf",
			ConfigMapName: appName,
		})
	}
	return volumes
}

func newPodContainers(name, nginxImage, appImage string, envVars map[string]string) []*Container {
	appPort := DefaultPort
	if nginxImage != "" {
		appPort = secondaryPort
	}

	c := []*Container{
		newAppContainer(name, appImage, envVars, appPort),
	}
	if nginxImage != "" {
		c = append(c, newNginxContainer(nginxImage))
	}
	return c
}

func NewPod(name, nginxImage, image string, a *app.App, envVars map[string]string, fs storage.Storage) *Pod {
	hasNginx := false
	hasNginx = nginxImage != ""

	ps := &Pod{
		Name:       name,
		Namespace:  a.Name,
		Containers: newPodContainers(name, nginxImage, image, envVars),
		Volumes:    newPodVolumes(a.Name, fs, hasNginx),
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
		"",
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

func NewRunner(name, slugURL string, imgs *Images, a *app.App, fs storage.Storage, cl *ContainerLimits, command ...string) *Pod {
	ps := NewPod(
		name,
		"",
		imgs.SlugRunner,
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
	ps.InitContainers = newInitContainers(slugURL, imgs.SlugStore, a, fs)
	return ps
}
