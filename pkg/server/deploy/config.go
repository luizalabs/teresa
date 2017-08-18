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
	ProcfileFileName       = "Procfile"
	teresaYamlFileNameTmpl = "teresa%s%s.yaml"
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

func (d *DeployConfigFiles) fillTeresaYaml(r io.Reader) error {
	d.TeresaYaml = new(TeresaYaml)
	if err := readFileFromTarBall(r, d.TeresaYaml); err != nil {
		return err
	}
	return validateTeresaYaml(d.TeresaYaml)
}

func readFileFromTarBall(r io.Reader, t interface{}) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, t)
}

func getDeployConfigFilesFromTarBall(tarBall io.ReadSeeker, processType string) (*DeployConfigFiles, error) {
	gReader, err := gzip.NewReader(tarBall)
	if err != nil {
		return nil, err
	}
	defer gReader.Close()

	tYamlFilename := fmt.Sprintf(teresaYamlFileNameTmpl, "", "")
	tYamlProcessTypeFileName := fmt.Sprintf(teresaYamlFileNameTmpl, "-", processType)
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

		if hdr.Name != tYamlFilename && hdr.Name != tYamlProcessTypeFileName && hdr.Name != ProcfileFileName {
			continue
		}

		if hdr.Name == tYamlFilename || hdr.Name == tYamlProcessTypeFileName {
			if hdr.Name == tYamlProcessTypeFileName {
				if err := deployFiles.fillTeresaYaml(tarReader); err != nil {
					return nil, err
				}
			} else if deployFiles.TeresaYaml == nil {
				if err := deployFiles.fillTeresaYaml(tarReader); err != nil {
					return nil, err
				}
			}
		} else {
			deployFiles.Procfile = make(map[string]string)
			if err := readFileFromTarBall(tarReader, deployFiles.Procfile); err != nil {
				return nil, err
			}
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
