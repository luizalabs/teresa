package spec

import (
	"strconv"
)

const (
	sharedVolumeName      = "shared-data"
	sharedVolumeMountPath = "/app"
)

type Labels map[string]string

type VolumeItem struct {
	Key  string
	Path string
}

type Volume struct {
	Name          string
	SecretName    string
	ConfigMapName string
	EmptyDir      bool
	Items         []VolumeItem
}

type Pod struct {
	Name           string
	Namespace      string
	Containers     []*Container
	Volumes        []*Volume
	InitContainers []*Container
	Labels         Labels
}

type PodBuilder struct {
	p             *Pod
	initContainer *Container
	appContainer  *Container
	sideCars      []*Container
}

func SwitchPortWithAppContainer(b *PodBuilder) {
	sideCarIdx := len(b.sideCars) - 1

	appPort, sideCarPort := b.appContainer.Ports[0].ContainerPort, b.sideCars[sideCarIdx].Ports[0].ContainerPort
	b.appContainer.Ports[0].ContainerPort, b.sideCars[sideCarIdx].Ports[0].ContainerPort = sideCarPort, appPort

	b.appContainer.Env["PORT"] = strconv.Itoa(int(b.appContainer.Ports[0].ContainerPort))
}

func shareVolumeWithAppContainer(name, path string, cn *Container, b *PodBuilder) {
	for _, c := range []*Container{b.appContainer, cn} {
		c.VolumeMounts = append(
			c.VolumeMounts,
			&VolumeMounts{Name: name, MountPath: path},
		)
	}
	b.p.Volumes = append(
		b.p.Volumes,
		&Volume{Name: name, EmptyDir: true},
	)
}

func ShareVolumeBetweenAppAndSideCar(name, path string) func(*PodBuilder) {
	return func(b *PodBuilder) {
		shareVolumeWithAppContainer(name, path, b.sideCars[len(b.sideCars)-1], b)
	}
}

func ShareVolumeBetweenAppAndInitContainer(name, path string) func(*PodBuilder) {
	return func(b *PodBuilder) {
		shareVolumeWithAppContainer(name, path, b.initContainer, b)
	}
}

func mountConfigMapInContainer(name, path, configMapName string, c *Container, b *PodBuilder) {
	c.VolumeMounts = append(
		c.VolumeMounts,
		&VolumeMounts{Name: name, MountPath: path, ReadOnly: true},
	)
	b.p.Volumes = append(
		b.p.Volumes,
		&Volume{Name: name, ConfigMapName: configMapName},
	)
}

func MountConfigMapInSideCar(name, path, configMapName string) func(*PodBuilder) {
	return func(b *PodBuilder) {
		mountConfigMapInContainer(name, path, configMapName, b.sideCars[len(b.sideCars)-1], b)
	}
}

func mountSecretInContainer(name, path, secretName string, c *Container, b *PodBuilder) {
	c.VolumeMounts = append(
		c.VolumeMounts,
		&VolumeMounts{Name: name, MountPath: path, ReadOnly: true},
	)
	b.p.Volumes = append(
		b.p.Volumes,
		&Volume{Name: name, SecretName: secretName},
	)
}

func MountSecretInInitContainer(name, path, secretName string) func(*PodBuilder) {
	return func(b *PodBuilder) {
		mountSecretInContainer(name, path, secretName, b.initContainer, b)
	}
}

func MountSecretInAppContainer(name, path, secretName string) func(*PodBuilder) {
	return func(b *PodBuilder) {
		mountSecretInContainer(name, path, secretName, b.appContainer, b)
	}
}

func MountSecretItemsInAppContainer(name, path, secretName string, items []string) func(*PodBuilder) {
	return func(b *PodBuilder) {
		if len(items) == 0 {
			return
		}
		b.appContainer.VolumeMounts = append(
			b.appContainer.VolumeMounts,
			&VolumeMounts{Name: name, MountPath: path, ReadOnly: true},
		)
		volItems := make([]VolumeItem, len(items))
		for i, item := range items {
			volItems[i].Key, volItems[i].Path = item, item
		}
		b.p.Volumes = append(
			b.p.Volumes,
			&Volume{Name: name, SecretName: secretName, Items: volItems},
		)
	}
}

func (b *PodBuilder) WithInitContainer(cn *Container, options ...func(*PodBuilder)) *PodBuilder {
	b.initContainer = cn
	for _, opt := range options {
		opt(b)
	}
	return b
}

func (b *PodBuilder) WithAppContainer(cn *Container, options ...func(*PodBuilder)) *PodBuilder {
	b.appContainer = cn
	for _, opt := range options {
		opt(b)
	}
	return b
}

func (b *PodBuilder) WithSideCar(cn *Container, options ...func(*PodBuilder)) *PodBuilder {
	b.sideCars = append(b.sideCars, cn)
	for _, opt := range options {
		opt(b)
	}
	return b
}

func (b *PodBuilder) WithLabels(lb Labels) *PodBuilder {
	for k, v := range lb {
		b.p.Labels[k] = v
	}
	return b
}

func (b *PodBuilder) Build() *Pod {
	b.p.Containers = make([]*Container, len(b.sideCars)+1)
	b.p.Containers[0] = b.appContainer
	for i := 0; i < len(b.sideCars); i++ {
		b.p.Containers[i+1] = b.sideCars[i]
	}
	if b.initContainer != nil {
		b.p.InitContainers = []*Container{b.initContainer}
	}
	return b.p
}

func NewPodBuilder(name, namespace string) *PodBuilder {
	p := &Pod{Name: name, Namespace: namespace, Labels: make(map[string]string)}
	return &PodBuilder{
		p:        p,
		sideCars: make([]*Container, 0),
	}
}
