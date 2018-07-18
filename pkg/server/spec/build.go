package spec

import (
	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

const (
	vlName = "storage-keys"
	vlPath = "/var/run/secrets/deis/objectstore/creds"
)

type BuildPodBuilder struct {
	name     string
	image    string
	tarPath  string
	slugDest string
	app      *app.App
	fs       storage.Storage
	cl       *ContainerLimits
}

func (b *BuildPodBuilder) newAppBuildContainer() *Container {
	env := map[string]string{
		"TAR_PATH":        b.tarPath,
		"PUT_PATH":        b.slugDest,
		"BUILDER_STORAGE": b.fs.Type(),
	}
	for _, ev := range b.app.EnvVars {
		env[ev.Key] = ev.Value
	}
	return NewContainerBuilder(b.name, b.image).
		WithEnv(env).
		WithEnv(b.fs.PodEnvVars()).
		WithSecrets(b.app.Secrets).
		WithLimits(b.cl.CPU, b.cl.Memory).
		Build()
}

func (b *BuildPodBuilder) newAppBuildPod(appContainer *Container) *Pod {
	return NewPodBuilder(b.name, b.app.Name).
		WithAppContainer(appContainer, MountSecretInAppContainer(vlName, vlPath, b.fs.K8sSecretName())).
		Build()
}

func (b *BuildPodBuilder) ForApp(a *app.App) *BuildPodBuilder {
	b.app = a
	return b
}

func (b *BuildPodBuilder) WithTarBallPath(path string) *BuildPodBuilder {
	b.tarPath = path
	return b
}

func (b *BuildPodBuilder) SendSlugTo(dest string) *BuildPodBuilder {
	b.slugDest = dest
	return b
}

func (b *BuildPodBuilder) WithLimits(cpu, memory string) *BuildPodBuilder {
	b.cl = &ContainerLimits{
		CPU:    cpu,
		Memory: memory,
	}
	return b
}

func (b *BuildPodBuilder) WithStorage(fs storage.Storage) *BuildPodBuilder {
	b.fs = fs
	return b
}

func (b *BuildPodBuilder) Build() *Pod {
	return b.newAppBuildPod(b.newAppBuildContainer())
}

func NewBuildPodBuilder(name, image string) *BuildPodBuilder {
	return &BuildPodBuilder{name: name, image: image}
}
