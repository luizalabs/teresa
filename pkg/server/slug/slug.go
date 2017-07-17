package slug

var (
	ProtectedEnvVars = [...]string{
		"SLUG_URL",
		"PORT",
		"DEIS_DEBUG",
		"BUILDER_STORAGE",
		"APP",
	}
)
