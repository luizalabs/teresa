package spec

import (
	"fmt"
	"path"

	"github.com/pkg/errors"
)

const (
	cloudSQLProxyDefaultCPULimit    = "100m"
	cloudSQLProxyDefaultMemoryLimit = "256Mi"
)

type CloudSQLProxy struct {
	Instances      string
	CredentialFile string `yaml:"credentialFile"`
	Image          string `yaml:"-"`
}

func NewCloudSQLProxy(img string, t *TeresaYaml) (*CloudSQLProxy, error) {
	if t == nil {
		return nil, nil
	}
	v, found := t.SideCars["cloudsql-proxy"]
	if !found {
		return nil, nil
	}
	csp := new(CloudSQLProxy)
	if err := v.Unmarshal(csp); err != nil {
		return nil, errors.Wrap(err, "failed to build cloudsql proxy")
	}
	csp.Image = img
	return csp, nil
}

func NewCloudSQLProxyContainer(csp *CloudSQLProxy) *Container {
	args := newCloudSQLProxyContainerArgs(csp)
	mpath := newCloudSQLProxyContainerMountPath(csp)
	return NewContainerBuilder("cloudsql-proxy", csp.Image).
		WithCommand([]string{"/cloud_sql_proxy"}).
		WithArgs(args).
		WithLimits(cloudSQLProxyDefaultCPULimit, cloudSQLProxyDefaultMemoryLimit).
		WithVolumeMount(AppSecretName, mpath, path.Base(mpath)).
		Build()
}

func newCloudSQLProxyContainerArgs(csp *CloudSQLProxy) []string {
	return []string{
		fmt.Sprintf("-instances=%s", csp.Instances),
		fmt.Sprintf("-credential_file=%s", newCloudSQLProxyContainerMountPath(csp)),
	}
}

func newCloudSQLProxyContainerMountPath(csp *CloudSQLProxy) string {
	return path.Join("/secrets/cloudsql", path.Base(csp.CredentialFile))
}
