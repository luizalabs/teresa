package spec

import (
	"strconv"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

const (
	DefaultPort                = 5000
	SlugAnnotation             = "teresa.io/slug"
	defaultDrainTimeoutSeconds = 10
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

type TeresaYaml struct {
	HealthCheck   *HealthCheck   `yaml:"healthCheck,omitempty"`
	RollingUpdate *RollingUpdate `yaml:"rollingUpdate,omitempty"`
	Lifecycle     *Lifecycle     `yaml:"lifecycle,omitempty"`
}

type Deploy struct {
	Pod
	TeresaYaml
	RevisionHistoryLimit int
	Description          string
	SlugURL              string
}

func NewDeploy(image, description, slugURL string, rhl int, a *app.App, tYaml *TeresaYaml, fs storage.Storage) *Deploy {
	ps := NewPod(
		a.Name,
		image,
		a,
		map[string]string{
			"APP":             a.Name,
			"PORT":            strconv.Itoa(DefaultPort),
			"SLUG_URL":        slugURL,
			"BUILDER_STORAGE": fs.Type(),
		},
		fs,
	)
	ps.Args = []string{"start", a.ProcessType}

	ds := &Deploy{
		Description:          description,
		SlugURL:              slugURL,
		Pod:                  *ps,
		RevisionHistoryLimit: rhl,
	}

	if tYaml != nil {
		ds.TeresaYaml = TeresaYaml{
			HealthCheck:   tYaml.HealthCheck,
			RollingUpdate: tYaml.RollingUpdate,
			Lifecycle:     tYaml.Lifecycle,
		}
	}

	if ds.Lifecycle == nil {
		ds.Lifecycle = &Lifecycle{
			PreStop: &PreStop{DrainTimeoutSeconds: defaultDrainTimeoutSeconds},
		}
	}

	return ds
}
