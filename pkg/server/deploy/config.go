package deploy

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

const (
	maxDrainTimeoutSeconds = 30
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

type Procfile map[string]string

type DeployConfigFiles struct {
	TeresaYaml *TeresaYaml
	Procfile   Procfile
}

const (
	TeresaYamlFileName = "teresa.yaml"
	ProcfileFileName   = "Procfile"
)

func readFileFromTarBall(r io.Reader, t interface{}) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, t)
}

func getDeployConfigFilesFromTarBall(tarBall io.ReadSeeker) (*DeployConfigFiles, error) {
	gReader, err := gzip.NewReader(tarBall)
	if err != nil {
		return nil, err
	}
	defer gReader.Close()

	deployFiles := new(DeployConfigFiles)
	tarReader := tar.NewReader(gReader)
	for {
		hdr, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if hdr.Name != TeresaYamlFileName && hdr.Name != ProcfileFileName {
			continue
		}

		if hdr.Name == TeresaYamlFileName {
			deployFiles.TeresaYaml = new(TeresaYaml)
			if err := readFileFromTarBall(tarReader, deployFiles.TeresaYaml); err != nil {
				return nil, err
			}
			if err := validateTeresaYaml(deployFiles.TeresaYaml); err != nil {
				return nil, err
			}
		} else {
			deployFiles.Procfile = make(map[string]string)
			if err := readFileFromTarBall(tarReader, deployFiles.Procfile); err != nil {
				return nil, err
			}
		}

		if deployFiles.TeresaYaml != nil && deployFiles.Procfile != nil {
			return deployFiles, nil
		}
	}

	return deployFiles, nil
}

func validateTeresaYaml(tYaml *TeresaYaml) error {
	if tYaml.Lifecycle != nil && tYaml.Lifecycle.PreStop != nil {
		if tYaml.Lifecycle.PreStop.DrainTimeoutSeconds > maxDrainTimeoutSeconds || tYaml.Lifecycle.PreStop.DrainTimeoutSeconds < 0 {
			return fmt.Errorf("Invalid drainTimeoutSeconds: %d", tYaml.Lifecycle.PreStop.DrainTimeoutSeconds)
		}
	}
	return nil
}
