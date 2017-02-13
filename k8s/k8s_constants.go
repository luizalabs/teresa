package k8s

const (
	slugBuilderName  = "deis-slugbuilder"
	slugBuilderImage = "luizalabs/slugbuilder:git-923c9f8"
	slugRunnerImage  = "luizalabs/slugrunner:git-044f85c"
	tarPath          = "TAR_PATH"
	putPath          = "PUT_PATH"
	debugKey         = "DEIS_DEBUG"
	builderStorage   = "BUILDER_STORAGE"
	objectStore      = "s3-storage"
	objectStorePath  = "/var/run/secrets/deis/objectstore/creds"
)
