package spec

const (
	DefaultPort                = 5000
	secondaryPort              = 6000
	SlugAnnotation             = "teresa.io/slug"
	defaultDrainTimeoutSeconds = 10
	DefaultExternalPort        = 80
)

type HealthCheckProbe struct {
	FailureThreshold    int32  `yaml:"failureThreshold"`
	InitialDelaySeconds int32  `yaml:"initialDelaySeconds"`
	PeriodSeconds       int32  `yaml:"periodSeconds"`
	SuccessThreshold    int32  `yaml:"successThreshold"`
	TimeoutSeconds      int32  `yaml:"timeoutSeconds"`
	Path                string `yaml:"path"`
}

type HealthCheck struct {
	Liveness  *HealthCheckProbe
	Readiness *HealthCheckProbe
}

type RollingUpdate struct {
	MaxSurge       string `yaml:"maxSurge,omitempty"`
	MaxUnavailable string `yaml:"maxUnavailable,omitempty"`
}

type PreStop struct {
	DrainTimeoutSeconds int `yaml:"drainTimeoutSeconds,omitempty"`
}

type Lifecycle struct {
	PreStop *PreStop `yaml:"preStop,omitempty"`
}

type CronArgs struct {
	Schedule string `yaml:"schedule",omitempty"`
}

type TeresaYaml struct {
	HealthCheck   *HealthCheck       `yaml:"healthCheck,omitempty"`
	RollingUpdate *RollingUpdate     `yaml:"rollingUpdate,omitempty"`
	Lifecycle     *Lifecycle         `yaml:"lifecycle,omitempty"`
	Cron          *CronArgs          `yaml:"cron,omitempty"`
	SideCars      map[string]RawData `yaml:"sidecars,omitempty"`
}

type TeresaYamlV2 struct {
	Applications map[string]*TeresaYaml `yaml:"applications,omitempty"`
}

type Deploy struct {
	Pod
	TeresaYaml
	RevisionHistoryLimit int
	Description          string
	SlugURL              string
	MatchLabels          Labels
}

type DeployBuilder struct {
	d *Deploy
}

type RawData struct {
	Fn func(interface{}) error
}

func (r *RawData) UnmarshalYAML(unmarshal func(interface{}) error) error {
	r.Fn = unmarshal
	return nil
}

func (r *RawData) Unmarshal(v interface{}) error {
	return r.Fn(v)
}

func (b *DeployBuilder) WithMatchLabels(lb Labels) *DeployBuilder {
	for k, v := range lb {
		b.d.MatchLabels[k] = v
	}
	return b
}

func (b *DeployBuilder) WithTeresaYaml(ty *TeresaYaml) *DeployBuilder {
	if ty != nil {
		b.d.TeresaYaml = *ty
	}
	return b
}

func (b *DeployBuilder) WithRevisionHistoryLimit(rhl int) *DeployBuilder {
	b.d.RevisionHistoryLimit = rhl
	return b
}

func (b *DeployBuilder) WithDescription(desc string) *DeployBuilder {
	b.d.Description = desc
	return b
}

func (b *DeployBuilder) WithPod(p *Pod) *DeployBuilder {
	b.d.Pod = *p
	return b
}

func (b *DeployBuilder) Build() *Deploy {
	if b.d.Lifecycle == nil {
		b.d.Lifecycle = &Lifecycle{
			PreStop: &PreStop{DrainTimeoutSeconds: defaultDrainTimeoutSeconds},
		}
	}
	return b.d
}

func NewDeployBuilder(slugURL string) *DeployBuilder {
	d := &Deploy{
		SlugURL:     slugURL,
		MatchLabels: make(Labels),
	}
	return &DeployBuilder{d: d}
}
