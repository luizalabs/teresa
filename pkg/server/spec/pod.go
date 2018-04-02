package spec

import (
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

const (
	slugVolumeName        = "slug"
	slugVolumeMountPath   = "/slug"
	sharedVolumeName      = "shared-data"
	sharedVolumeMountPath = "/app"
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
		volumes = append(volumes,
			&Volume{Name: "nginx-conf", ConfigMapName: appName},
			&Volume{Name: sharedVolumeName, EmptyDir: true},
		)
	}
	return volumes
}

func newPodContainers(name, nginxImage, appImage string, envVars map[string]string, secrets []string) []*Container {
	appPort := DefaultPort
	if nginxImage != "" {
		appPort = secondaryPort
	}

	c := []*Container{
		newAppContainer(name, appImage, envVars, appPort, secrets),
	}
	if nginxImage != "" {
		c[0].VolumeMounts = []*VolumeMounts{
			&VolumeMounts{
				Name:      sharedVolumeName,
				MountPath: sharedVolumeMountPath,
			},
		}
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
		Containers: newPodContainers(name, nginxImage, image, envVars, a.Secrets),
		Volumes:    newPodVolumes(a.Name, fs, hasNginx),
	}

	for _, e := range a.EnvVars {
		for i := range ps.Containers {
			ps.Containers[i].Env[e.Key] = e.Value
		}
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
