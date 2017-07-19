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

const TeresaYamlFileName = "teresa.yaml"

func getTeresaYamlFromTarBall(tarBall io.ReadSeeker) (*TeresaYaml, error) {
	gReader, err := gzip.NewReader(tarBall)
	if err != nil {
		return nil, err
	}
	defer gReader.Close()

	tarReader := tar.NewReader(gReader)
	for {
		hdr, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if hdr.Name != TeresaYamlFileName {
			continue
		}

		b, err := ioutil.ReadAll(tarReader)
		if err != nil {
			return nil, err
		}

		teresaYaml := new(TeresaYaml)
		if err = yaml.Unmarshal(b, teresaYaml); err != nil {
			return nil, err
		}

		if err := validateTeresaYaml(teresaYaml); err != nil {
			return nil, err
		}

		return teresaYaml, nil
	}

	return nil, nil
}

func validateTeresaYaml(tYaml *TeresaYaml) error {
	if tYaml.Lifecycle != nil && tYaml.Lifecycle.PreStop != nil {
		if tYaml.Lifecycle.PreStop.DrainTimeoutSeconds > maxDrainTimeoutSeconds || tYaml.Lifecycle.PreStop.DrainTimeoutSeconds < 0 {
			return fmt.Errorf("Invalid drainTimeoutSeconds: %d", tYaml.Lifecycle.PreStop.DrainTimeoutSeconds)
		}
	}
	return nil
}
