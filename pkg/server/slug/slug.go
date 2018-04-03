package slug

var (
	ProtectedEnvVars = [...]string{
		"PYTHONPATH",
		"SLUG_URL",
		"PORT",
		"DEIS_DEBUG",
		"BUILDER_STORAGE",
		"APP",
		"SLUG_DIR",
		"NGINX_PORT",
		"NGINX_BACKEND",
	}
)
