package validation

import (
	"regexp"
)

var (
	envVarNameRegexp = regexp.MustCompile(`^[-._a-zA-Z][-._\w]*$`)
	ProtectedEnvVars = map[string]bool{
		"PYTHONPATH":      true,
		"SLUG_URL":        true,
		"PORT":            true,
		"DEIS_DEBUG":      true,
		"BUILDER_STORAGE": true,
		"APP":             true,
		"SLUG_DIR":        true,
		"NGINX_PORT":      true,
		"NGINX_BACKEND":   true,
	}
)

func IsEnvVarName(name string) bool {
	return envVarNameRegexp.MatchString(name)
}

func IsProtectedEnvVar(name string) bool {
	_, found := ProtectedEnvVars[name]
	return found
}
