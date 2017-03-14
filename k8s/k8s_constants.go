package k8s

const (
	slugBuilderName  = "deis-slugbuilder"
	slugBuilderImage = "teresak8s/slugbuilder:v2.4.9"
	slugRunnerImage  = "teresak8s/slugrunner:v2.2.4"
	tarPath          = "TAR_PATH"
	putPath          = "PUT_PATH"
	debugKey         = "DEIS_DEBUG"
	builderStorage   = "BUILDER_STORAGE"
	objectStore      = "s3-storage"
	objectStorePath  = "/var/run/secrets/deis/objectstore/creds"
)
