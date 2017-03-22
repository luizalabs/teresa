package k8s

const (
	slugBuilderName  = "deis-slugbuilder"
	slugBuilderImage = "luizalabs/slugbuilder:v2.4.9"
	slugRunnerImage  = "luizalabs/slugrunner:v2.2.4"
	tarPath          = "TAR_PATH"
	putPath          = "PUT_PATH"
	debugKey         = "DEIS_DEBUG"
	builderStorage   = "BUILDER_STORAGE"
	objectStore      = "s3-storage"
	objectStorePath  = "/var/run/secrets/deis/objectstore/creds"
)
