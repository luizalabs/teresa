package deploy

import (
	"fmt"
	"strconv"

	"github.com/luizalabs/teresa/pkg/server/app"
	st "github.com/luizalabs/teresa/pkg/server/storage"
)

const (
	DefaultPort                = 5000
	defaultDrainTimeoutSeconds = 10
)

type ContainerLimits struct {
	CPU    string
	Memory string
}

type PodVolumeMountsSpec struct {
	Name      string
	MountPath string
	ReadOnly  bool
}

type PodVolumeSpec struct {
	Name       string
	SecretName string
}

type PodSpec struct {
	Name            string
	Namespace       string
	Image           string
	ContainerLimits *ContainerLimits
	Env             map[string]string
	VolumeMounts    []*PodVolumeMountsSpec
	Volume          []*PodVolumeSpec
	Args            []string
}

type DeploySpec struct {
	PodSpec
	TeresaYaml
	RevisionHistoryLimit int
	Description          string
	SlugURL              string
}

func newPodSpec(name, image string, a *app.App, envVars map[string]string, fileStorage st.Storage) *PodSpec {
	ps := &PodSpec{
		Name:      name,
		Namespace: a.Name,
		Image:     image,
		VolumeMounts: []*PodVolumeMountsSpec{{
			Name:      "storage-keys",
			MountPath: "/var/run/secrets/deis/objectstore/creds",
			ReadOnly:  true,
		}},
		Volume: []*PodVolumeSpec{{
			Name:       "storage-keys",
			SecretName: fileStorage.K8sSecretName(),
		}},
		Env: envVars,
	}
	for _, e := range a.EnvVars {
		ps.Env[e.Key] = e.Value
	}
	for k, v := range fileStorage.PodEnvVars() {
		ps.Env[k] = v
	}
	return ps
}

func newBuildSpec(a *app.App, deployId, tarBallLocation, buildDest string, fileStorage st.Storage, opts *Options) *PodSpec {
	ps := newPodSpec(
		fmt.Sprintf("build-%s", deployId),
		opts.SlugBuilderImage,
		a,
		map[string]string{
			"TAR_PATH":        tarBallLocation,
			"PUT_PATH":        buildDest,
			"BUILDER_STORAGE": fileStorage.Type(),
		},
		fileStorage,
	)
	ps.ContainerLimits = &ContainerLimits{
		CPU:    opts.BuildLimitCPU,
		Memory: opts.BuildLimitMemory,
	}
	return ps
}

func newDeploySpec(a *app.App, tYaml *TeresaYaml, fileStorage st.Storage, description, slugURL, processType string, opts *Options) *DeploySpec {
	ps := newPodSpec(
		a.Name,
		opts.SlugRunnerImage,
		a,
		map[string]string{
			"APP":             a.Name,
			"PORT":            strconv.Itoa(DefaultPort),
			"SLUG_URL":        slugURL,
			"BUILDER_STORAGE": fileStorage.Type(),
		},
		fileStorage,
	)
	ps.Args = []string{"start", processType}

	ds := &DeploySpec{
		Description:          description,
		SlugURL:              slugURL,
		PodSpec:              *ps,
		RevisionHistoryLimit: opts.RevisionHistoryLimit,
	}

	if tYaml != nil {
		ds.TeresaYaml = TeresaYaml{
			HealthCheck:   tYaml.HealthCheck,
			RollingUpdate: tYaml.RollingUpdate,
			Lifecycle:     tYaml.Lifecycle,
		}
	}

	setDefaultLifecycle(ds)

	return ds
}

func newRunCommandSpec(a *app.App, deployId, command, slugURL string, fileStorage st.Storage, opts *Options) *PodSpec {
	ps := newPodSpec(
		fmt.Sprintf("release-%s-%s", a.Name, deployId),
		opts.SlugRunnerImage,
		a,
		map[string]string{
			"APP":             a.Name,
			"PORT":            strconv.Itoa(DefaultPort),
			"SLUG_URL":        slugURL,
			"BUILDER_STORAGE": fileStorage.Type(),
		},
		fileStorage,
	)
	ps.Args = []string{"start", command}
	ps.ContainerLimits = &ContainerLimits{
		CPU:    opts.BuildLimitCPU,
		Memory: opts.BuildLimitMemory,
	}
	return ps
}

func setDefaultLifecycle(ds *DeploySpec) {
	if ds.Lifecycle == nil {
		ds.Lifecycle = &Lifecycle{
			PreStop: &PreStop{DrainTimeoutSeconds: defaultDrainTimeoutSeconds},
		}
	}
}
