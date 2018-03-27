package spec

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/luizalabs/teresa/pkg/server/app"
	"github.com/luizalabs/teresa/pkg/server/storage"
)

const (
	nginxConfTmplDir = "/etc/nginx/template/"
	nginxConfDir     = "/etc/nginx/"
	nginxConfFile    = "nginx.conf"
	nginxArgTmpl     = "envsubst '%s' < %s%s > %s%s && nginx -g 'daemon off;'"
	nginxBackendTmpl = "http://localhost:%d"
)

type ContainerLimits struct {
	CPU    string
	Memory string
}

type VolumeMounts struct {
	Name      string
	MountPath string
	ReadOnly  bool
}

type Port struct {
	Name          string
	ContainerPort int32
}

type Container struct {
	Name            string
	Image           string
	ContainerLimits *ContainerLimits
	Env             map[string]string
	VolumeMounts    []*VolumeMounts
	Command         []string
	Args            []string
	Ports           []Port
	Secrets         []string
}

func newSlugVolumeMount() *VolumeMounts {
	return &VolumeMounts{
		Name:      slugVolumeName,
		MountPath: slugVolumeMountPath,
	}
}

func newStorageKeyVolumeMount() *VolumeMounts {
	return &VolumeMounts{
		Name:      "storage-keys",
		MountPath: "/var/run/secrets/deis/objectstore/creds",
		ReadOnly:  true,
	}
}

func newInitContainers(slugURL, image string, a *app.App, fs storage.Storage) []*Container {
	ic := &Container{
		Name:  "slugstore",
		Image: image,
		Env: map[string]string{
			"BUILDER_STORAGE": fs.Type(),
			"SLUG_URL":        slugURL,
			"SLUG_DIR":        slugVolumeMountPath,
		},
		VolumeMounts: []*VolumeMounts{
			newStorageKeyVolumeMount(),
			newSlugVolumeMount(),
		},
	}
	for k, v := range fs.PodEnvVars() {
		ic.Env[k] = v
	}
	return []*Container{ic}
}

func newNginxVolumeMounts() []*VolumeMounts {
	return []*VolumeMounts{
		&VolumeMounts{
			Name:      "nginx-conf",
			MountPath: nginxConfTmplDir,
			ReadOnly:  true,
		},
		&VolumeMounts{
			Name:      sharedVolumeName,
			MountPath: sharedVolumeMountPath,
		},
	}
}

func newNginxContainer(image string) *Container {
	port := strconv.Itoa(DefaultPort)
	backend := fmt.Sprintf(nginxBackendTmpl, secondaryPort)
	env := map[string]string{
		"NGINX_PORT":    port,
		"NGINX_BACKEND": backend,
	}
	args := newNginxContainerArgs(env)

	return &Container{
		Name:    "nginx",
		Image:   image,
		Command: []string{"/bin/sh"},
		Args:    []string{"-c", args},
		Ports: []Port{{
			Name:          "nginx",
			ContainerPort: int32(DefaultPort),
		}},
		Env:          env,
		VolumeMounts: newNginxVolumeMounts(),
	}
}

func newNginxContainerArgs(env map[string]string) string {
	tmp := make([]string, len(env))
	var i int
	for key, _ := range env {
		tmp[i] = fmt.Sprintf("$%s", key)
		i++
	}
	sort.Strings(tmp)

	args := fmt.Sprintf(
		nginxArgTmpl,
		strings.Join(tmp, " "),
		nginxConfTmplDir,
		nginxConfFile,
		nginxConfDir,
		nginxConfFile,
	)
	return args
}

func newAppContainer(name, image string, envVars map[string]string, port int, secrets []string) *Container {
	return &Container{
		Name:    name,
		Image:   image,
		Env:     envVars,
		Secrets: secrets,
		Ports: []Port{{
			Name:          "app",
			ContainerPort: int32(port),
		}},
	}
}
