package spec

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewCloudSQLProxy(t *testing.T) {
	img := "image"
	want := &CloudSQLProxy{
		Instances:      "instances",
		CredentialFile: "file",
		Image:          img,
	}
	fn := func(v interface{}) error {
		v.(*CloudSQLProxy).Instances = want.Instances
		v.(*CloudSQLProxy).CredentialFile = want.CredentialFile
		return nil
	}
	ty := &TeresaYaml{
		SideCars: map[string]RawData{
			"cloudsql-proxy": {fn},
		},
	}

	csp, err := NewCloudSQLProxy(img, ty)
	if err != nil {
		t.Fatal("got unexpected error:", err)
	}
	if !reflect.DeepEqual(csp, want) {
		t.Errorf("got %v; want %v", csp, want)
	}
}

func TestNewCloudSQLProxyError(t *testing.T) {
	fn := func(v interface{}) error {
		return errors.New("test")
	}
	ty := &TeresaYaml{
		SideCars: map[string]RawData{
			"cloudsql-proxy": {fn},
		},
	}

	if _, err := NewCloudSQLProxy("test", ty); err == nil {
		t.Error("got nil; want error")
	}
}

func TestNewCloudSQLProxyNil(t *testing.T) {
	ty := &TeresaYaml{
		SideCars: map[string]RawData{
			"test": RawData{},
		},
	}

	csp, err := NewCloudSQLProxy("test", ty)
	if err != nil {
		t.Fatal("got unexpected error:", err)
	}
	if csp != nil {
		t.Errorf("got %v; want nil", csp)
	}
}

func TestNewCloudSQLProxyContainer(t *testing.T) {
	csp := &CloudSQLProxy{
		Instances:      "instances",
		CredentialFile: "file",
		Image:          "image",
	}
	want := &Container{
		Name: "cloudsql-proxy",
		ContainerLimits: &ContainerLimits{
			CPU:    cloudSQLProxyDefaultCPULimit,
			Memory: cloudSQLProxyDefaultMemoryLimit,
		},
		Image:   csp.Image,
		Command: []string{"/cloud_sql_proxy"},
		Args: []string{
			"-instances=instances",
			"-credential_file=/secrets/cloudsql/file",
		},
		VolumeMounts: []*VolumeMounts{
			{
				Name:      AppSecretName,
				MountPath: "/secrets/cloudsql/file",
				SubPath:   "file",
				ReadOnly:  true,
			},
		},
		Env:     map[string]string{},
		Ports:   []Port{},
		Secrets: []string{},
	}

	cn := NewCloudSQLProxyContainer(csp)
	if !reflect.DeepEqual(cn, want) {
		t.Errorf("got %v; want %v", cn, want)
	}
}
