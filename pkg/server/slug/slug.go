package slug

var (
	ProtectedEnvVars = [...]string{
		"PYTHONPATH",
		"SLUG_URL",
		"PORT",
		"DEIS_DEBUG",
		"BUILDER_STORAGE",
		"APP",
	}
)
