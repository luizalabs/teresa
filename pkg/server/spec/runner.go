package spec

import (
	"strconv"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

const AppSecretName = "secrets"

type RunnerPodBuilder struct {
	name       string
	image      string
	initImage  string
	slugURL    string
	nginxImage string
	args       []string
	app        *app.App
	fs         storage.Storage
	cl         *ContainerLimits
	labels     Labels
	csp        *CloudSQLProxy
}

func (b *RunnerPodBuilder) newAppRunnerContainer() *Container {
	env := map[string]string{
		"APP":      b.app.Name,
		"SLUG_URL": b.slugURL,
		"SLUG_DIR": slugVolumeMountPath,
	}
	for _, ev := range b.app.EnvVars {
		env[ev.Key] = ev.Value
	}
	builder := NewContainerBuilder(b.name, b.image).
		WithEnv(env).
		WithSecrets(b.app.Secrets).
		WithArgs(b.args)

	if app.IsWebApp(b.app.ProcessType) {
		builder = builder.
			WithEnv(map[string]string{"PORT": strconv.Itoa(DefaultPort)}).
			ExposePort("http", DefaultPort)
	}
	if b.cl != nil {
		builder = builder.WithLimits(b.cl.CPU, b.cl.Memory)
	}
	return builder.Build()
}

func (b *RunnerPodBuilder) newAppRunnerPod(appContainer *Container) *Pod {
	init := NewInitContainer(b.initImage, b.slugURL, b.fs)
	mountSecretOpt := MountSecretInInitContainer(vlName, vlPath, b.fs.K8sSecretName())
	shareVolOpt := ShareVolumeBetweenAppAndInitContainer(slugVolumeName, slugVolumeMountPath)

	msc := MountSecretItemsInAppContainer(
		AppSecretName,
		app.SecretPath,
		app.TeresaAppSecrets,
		b.app.SecretFiles,
	)

	builder := NewPodBuilder(b.name, b.app.Name).
		WithAppContainer(appContainer, msc).
		WithLabels(b.labels).
		WithInitContainer(init, mountSecretOpt, shareVolOpt)

	if b.nginxImage != "" {
		nc := NewNginxContainer(b.nginxImage, b.app)
		nVol := ShareVolumeBetweenAppAndSideCar(sharedVolumeName, sharedVolumeMountPath)
		nCm := MountConfigMapInSideCar(nginxVolName, nginxConfTmplDir, b.app.Name)

		builder = builder.WithSideCar(nc, nVol, nCm, SwitchPortWithAppContainer)
	}

	if b.csp != nil {
		cn := NewCloudSQLProxyContainer(b.csp)
		builder = builder.WithSideCar(cn)
	}

	return builder.Build()
}

func (b *RunnerPodBuilder) ForApp(a *app.App) *RunnerPodBuilder {
	b.app = a
	return b
}

func (b *RunnerPodBuilder) WithSlug(url string) *RunnerPodBuilder {
	b.slugURL = url
	return b
}

func (b *RunnerPodBuilder) WithLimits(cpu, memory string) *RunnerPodBuilder {
	b.cl = &ContainerLimits{
		CPU:    cpu,
		Memory: memory,
	}
	return b
}

func (b *RunnerPodBuilder) WithStorage(fs storage.Storage) *RunnerPodBuilder {
	b.fs = fs
	return b
}

func (b *RunnerPodBuilder) WithArgs(args []string) *RunnerPodBuilder {
	b.args = args
	return b
}

func (b *RunnerPodBuilder) WithNginxSideCar(image string) *RunnerPodBuilder {
	b.nginxImage = image
	return b
}

func (b *RunnerPodBuilder) WithLabels(lb Labels) *RunnerPodBuilder {
	for k, v := range lb {
		b.labels[k] = v
	}
	return b
}

func (b *RunnerPodBuilder) Build() *Pod {
	return b.newAppRunnerPod(b.newAppRunnerContainer())
}

func (b *RunnerPodBuilder) WithCloudSQLProxySideCar(csp *CloudSQLProxy) *RunnerPodBuilder {
	b.csp = csp
	return b
}

func NewRunnerPodBuilder(name, image, initImage string) *RunnerPodBuilder {
	return &RunnerPodBuilder{
		name:      name,
		image:     image,
		initImage: initImage,
		labels:    make(map[string]string),
	}
}
