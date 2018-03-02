package deploy

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/luizalabs/teresa/pkg/server/spec"
	yaml "gopkg.in/yaml.v2"
)

const (
	ProcfileFileName       = "Procfile"
	teresaYamlFileNameTmpl = "teresa%s%s.yaml"
	maxDrainTimeoutSeconds = 30
	nginxConfFileName      = "nginx.conf"
)

type Procfile map[string]string

type DeployConfigFiles struct {
	TeresaYaml *spec.TeresaYaml
	Procfile   Procfile
	NginxConf  string
}

func (d *DeployConfigFiles) fillTeresaYaml(r io.Reader) error {
	d.TeresaYaml = new(spec.TeresaYaml)
	if err := readYAMLFromTarBall(r, d.TeresaYaml); err != nil {
		return err
	}
	return validateTeresaYaml(d.TeresaYaml)
}

func readFileFromTarBall(r io.Reader) (string, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func readYAMLFromTarBall(r io.Reader, t interface{}) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(b, t)
}

func isConfigFile(name string, confFiles ...string) bool {
	for _, cf := range confFiles {
		if name == cf {
			return true
		}
	}
	return false
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

		if !isConfigFile(hdr.Name, tYamlFilename, tYamlProcessTypeFileName, ProcfileFileName, nginxConfFileName) {
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
		} else if hdr.Name == nginxConfFileName {
			deployFiles.NginxConf, err = readFileFromTarBall(tarReader)
			if err != nil {
				return nil, err
			}
		} else {
			deployFiles.Procfile = make(map[string]string)
			if err := readYAMLFromTarBall(tarReader, deployFiles.Procfile); err != nil {
				return nil, err
			}
		}
	}

	return deployFiles, nil
}

func validateTeresaYaml(tYaml *spec.TeresaYaml) error {
	if tYaml.Lifecycle != nil && tYaml.Lifecycle.PreStop != nil {
		if tYaml.Lifecycle.PreStop.DrainTimeoutSeconds > maxDrainTimeoutSeconds || tYaml.Lifecycle.PreStop.DrainTimeoutSeconds < 0 {
			return fmt.Errorf("Invalid drainTimeoutSeconds: %d", tYaml.Lifecycle.PreStop.DrainTimeoutSeconds)
		}
	}
	return nil
}
