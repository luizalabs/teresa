package deploy

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/luizalabs/teresa/pkg/server/spec"
	yaml "gopkg.in/yaml.v2"
)

const (
	ProcfileFileName       = "Procfile"
	maxDrainTimeoutSeconds = 30
	nginxConfFileName      = "nginx.conf"
)

type Procfile map[string]string

type DeployConfigFiles struct {
	TeresaYaml *spec.TeresaYaml
	Procfile   Procfile
	NginxConf  string
}

func (d *DeployConfigFiles) fillTeresaYaml(r io.Reader, appName string) error {
	type T struct {
		Version           string `yaml:"version"`
		spec.TeresaYaml   `yaml:",inline"`
		spec.TeresaYamlV2 `yaml:",inline"`
	}
	tmp := new(T)
	if err := readYAMLFromTarBall(r, tmp); err != nil {
		return err
	}
	switch tmp.Version {
	case "v2":
		if d.TeresaYaml = tmp.TeresaYamlV2.Applications[appName]; d.TeresaYaml == nil {
			d.TeresaYaml = &spec.TeresaYaml{}
		}
	default:
		d.TeresaYaml = &tmp.TeresaYaml
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

func getDeployConfigFilesFromTarBall(tarBall io.ReadSeeker, appName, processType string) (*DeployConfigFiles, error) {
	gReader, err := gzip.NewReader(tarBall)
	if err != nil {
		return nil, err
	}
	defer gReader.Close()

	deployFiles := new(DeployConfigFiles)
	tarReader := tar.NewReader(gReader)
	names := newConfigFileNames(processType)
	for {
		hdr, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if _, found := names[hdr.Name]; !found {
			continue
		}
		if hdr.Name == nginxConfFileName {
			deployFiles.NginxConf, err = readFileFromTarBall(tarReader)
			if err != nil {
				return nil, err
			}
		} else if hdr.Name == ProcfileFileName {
			deployFiles.Procfile = make(map[string]string)
			if err := readYAMLFromTarBall(tarReader, deployFiles.Procfile); err != nil {
				return nil, err
			}
		} else {
			if deployFiles.TeresaYaml == nil || strings.Index(hdr.Name, processType) > 0 {
				if err := deployFiles.fillTeresaYaml(tarReader, appName); err != nil {
					return nil, err
				}
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

func newConfigFileNames(processType string) map[string]bool {
	m := map[string]bool{
		ProcfileFileName:  true,
		nginxConfFileName: true,
	}
	for _, ext := range []string{".yaml", ".yml"} {
		for _, item := range []string{"", processType} {
			k := strings.TrimSuffix(fmt.Sprintf("teresa-%s", item), "-")
			k = fmt.Sprintf("%s%s", k, ext)
			m[k] = true
		}
	}
	return m
}
