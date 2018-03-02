package spec

import (
	"strconv"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

const (
	DefaultPort                = 5000
	secondaryPort              = 6000
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

type CronArgs struct {
	Schedule string `yaml:"schedule",omitempty"`
}

type TeresaYaml struct {
	HealthCheck   *HealthCheck   `yaml:"healthCheck,omitempty"`
	RollingUpdate *RollingUpdate `yaml:"rollingUpdate,omitempty"`
	Lifecycle     *Lifecycle     `yaml:"lifecycle,omitempty"`
	Cron          *CronArgs      `yaml:"cron,omitempty"`
}

type Deploy struct {
	Pod
	TeresaYaml
	RevisionHistoryLimit int
	Description          string
	SlugURL              string
}

type SlugImages struct {
	Runner string
	Store  string
}

func NewDeploy(imgs *SlugImages, description, slugURL string, rhl int, a *app.App, tYaml *TeresaYaml, fs storage.Storage) *Deploy {
	ps := NewPod(
		a.Name,
		imgs.Runner,
		a,
		map[string]string{
			"APP":      a.Name,
			"PORT":     strconv.Itoa(DefaultPort),
			"SLUG_URL": slugURL,
			"SLUG_DIR": slugVolumeMountPath,
		},
		fs,
	)
	ps.Containers[0].Args = []string{"start", a.ProcessType}
	ps.Containers[0].VolumeMounts = []*VolumeMounts{newSlugVolumeMount()}
	ps.InitContainers = newInitContainers(slugURL, imgs.Store, a, fs)

	ds := &Deploy{
		Description:          description,
		SlugURL:              slugURL,
		Pod:                  *ps,
		RevisionHistoryLimit: rhl,
	}

	if tYaml != nil {
		ds.TeresaYaml = *tYaml
	}

	if ds.Lifecycle == nil {
		ds.Lifecycle = &Lifecycle{
			PreStop: &PreStop{DrainTimeoutSeconds: defaultDrainTimeoutSeconds},
		}
	}

	return ds
}
