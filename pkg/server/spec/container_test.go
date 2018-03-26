package spec

import (
	"fmt"
	"testing"
)

func TestNewNginxContainerArgs(t *testing.T) {
	env := map[string]string{"key1": "value1", "key2": "value2"}
	keys := "$key1 $key2"
	want := fmt.Sprintf(
		nginxArgTmpl,
		keys,
		nginxConfTmplDir,
		nginxConfFile,
		nginxConfDir,
		nginxConfFile,
	)

	got := newNginxContainerArgs(env)

	if got != want {
		t.Errorf("got %s; want %s", got, want)
	}
}
