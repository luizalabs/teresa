package spec

import (
	"fmt"
	"testing"

	"github.com/luizalabs/teresa/pkg/server/app"
)

func TestNewNginxContainerArgs(t *testing.T) {
	env := map[string]string{"key1": "value1", "key2": "value2"}
	keys := "$key1 $key2"
	want := fmt.Sprintf(
		nginxArgTmpl,
		keys,
		nginxConfTmplDir,
		NginxConfFile,
		nginxConfDir,
		NginxConfFile,
	)

	if got := newNginxContainerArgs(env); got != want {
		t.Errorf("got %s; want %s", got, want)
	}
}

func TestNewNginxContainer(t *testing.T) {
	a := &app.App{
		EnvVars: []*app.EnvVar{&app.EnvVar{Key: "key1", Value: "value1"}},
	}
	expectedImage := "nginx"
	c := NewNginxContainer(expectedImage, a)
	if actual := c.Image; actual != expectedImage {
		t.Errorf("expected %s, got %s", expectedImage, actual)
	}
	expected := "/bin/sh"
	if actual := c.Command[0]; actual != expected {
		t.Errorf("expected %s, got %s", expected, actual)
	}
	if actual := c.Env["key1"]; actual != "value1" {
		t.Errorf("expected value1, got %s", actual)
	}
	if actual := c.Ports[0].ContainerPort; actual != int32(secondaryPort) {
		t.Errorf("expected %d, got %d", secondaryPort, actual)
	}
	if actual := c.ContainerLimits.CPU; actual != nginxDefaultCPULimit {
		t.Errorf("expected %s, got %s", nginxDefaultCPULimit, actual)
	}
}
