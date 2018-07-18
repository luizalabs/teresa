package spec

type ContainerLimits struct {
	CPU    string
	Memory string
}

type VolumeMounts struct {
	Name      string
	MountPath string
	ReadOnly  bool
}

type Port struct {
	Name          string
	ContainerPort int32
}

type Container struct {
	Name            string
	Image           string
	ContainerLimits *ContainerLimits
	Env             map[string]string
	VolumeMounts    []*VolumeMounts
	Command         []string
	Args            []string
	Ports           []Port
	Secrets         []string
}

type ContainerBuilder struct {
	c *Container
}

func (b *ContainerBuilder) WithCommand(cmd []string) *ContainerBuilder {
	b.c.Command = cmd
	return b
}

func (b *ContainerBuilder) WithArgs(args []string) *ContainerBuilder {
	b.c.Args = args
	return b
}

func (b *ContainerBuilder) WithEnv(ev map[string]string) *ContainerBuilder {
	for k, v := range ev {
		b.c.Env[k] = v
	}
	return b
}

func (b *ContainerBuilder) WithSecrets(s []string) *ContainerBuilder {
	b.c.Secrets = append(b.c.Secrets, s...)
	return b
}

func (b *ContainerBuilder) WithLimits(cpu, memory string) *ContainerBuilder {
	b.c.ContainerLimits = &ContainerLimits{
		CPU:    cpu,
		Memory: memory,
	}
	return b
}

func (b *ContainerBuilder) ExposePort(name string, port int) *ContainerBuilder {
	b.c.Ports = append(b.c.Ports, Port{Name: name, ContainerPort: int32(port)})
	return b
}

func (b *ContainerBuilder) Build() *Container {
	return b.c
}

func NewContainerBuilder(name, image string) *ContainerBuilder {
	return &ContainerBuilder{
		c: &Container{
			Name:         name,
			Image:        image,
			Env:          make(map[string]string),
			Secrets:      make([]string, 0),
			Ports:        make([]Port, 0),
			VolumeMounts: make([]*VolumeMounts, 0),
		},
	}
}
