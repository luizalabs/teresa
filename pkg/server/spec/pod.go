package spec

import (
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

type ContainerLimits struct {
	CPU    string
	Memory string
}

type VolumeMounts struct {
	Name      string
	MountPath string
	ReadOnly  bool
}

type Volume struct {
	Name       string
	SecretName string
}

type Container struct {
	Name            string
	Namespace       string
	Image           string
	ContainerLimits *ContainerLimits
	Env             map[string]string
	VolumeMounts    []*VolumeMounts
	Args            []string
}

type Pod struct {
	Container
	Volumes []*Volume
}

func NewPod(name, image string, a *app.App, envVars map[string]string, fs storage.Storage) *Pod {
	ps := &Pod{
		Container: Container{
			Name:      name,
			Namespace: a.Name,
			Image:     image,
			VolumeMounts: []*VolumeMounts{{
				Name:      "storage-keys",
				MountPath: "/var/run/secrets/deis/objectstore/creds",
				ReadOnly:  true,
			}},
			Env: envVars,
		},
		Volumes: []*Volume{{
			Name:       "storage-keys",
			SecretName: fs.K8sSecretName(),
		}},
	}
	for _, e := range a.EnvVars {
		ps.Env[e.Key] = e.Value
	}
	for k, v := range fs.PodEnvVars() {
		ps.Env[k] = v
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
	ps.ContainerLimits = cl
	return ps
}

func NewRunner(name, slugURL, image string, a *app.App, fs storage.Storage, cl *ContainerLimits, command ...string) *Pod {
	ps := NewPod(
		name,
		image,
		a,
		map[string]string{
			"APP":             a.Name,
			"SLUG_URL":        slugURL,
			"BUILDER_STORAGE": fs.Type(),
		},
		fs,
	)
	ps.Args = command
	ps.ContainerLimits = cl
	return ps
}
