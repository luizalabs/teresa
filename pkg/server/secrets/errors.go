package secrets

import "errors"

var (
	ErrMissingKey          = errors.New("Missing auth key")
	ErrMissingNamespaceEnv = errors.New("Missing 'NAMESPACE' environment variable")
)
