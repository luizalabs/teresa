package spec

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/luizalabs/teresa/pkg/server/app"
)

const (
	NginxConfFile           = "nginx.conf"
	nginxConfTmplDir        = "/etc/nginx/template/"
	nginxConfDir            = "/etc/nginx/"
	nginxVolName            = "nginx-conf"
	nginxArgTmpl            = "envsubst '%s' < %s%s > %s%s && nginx -g 'daemon off;'"
	nginxBackendTmpl        = "http://localhost:%d"
	nginxDefaultCPULimit    = "100m"
	nginxDefaultMemoryLimit = "256Mi"
)

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
		NginxConfFile,
		nginxConfDir,
		NginxConfFile,
	)
	return args
}

func NewNginxContainer(image string, a *app.App) *Container {
	env := map[string]string{
		"NGINX_PORT":    strconv.Itoa(DefaultPort),
		"NGINX_BACKEND": fmt.Sprintf("http://localhost:%d", secondaryPort),
	}
	args := newNginxContainerArgs(env)
	for _, e := range a.EnvVars {
		env[e.Key] = e.Value
	}
	return NewContainerBuilder("nginx", image).
		WithCommand([]string{"/bin/sh"}).
		WithArgs([]string{"-c", args}).
		WithEnv(env).
		WithLimits(nginxDefaultCPULimit, nginxDefaultMemoryLimit).
		ExposePort("nginx", secondaryPort).
		Build()
}
