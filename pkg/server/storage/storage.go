package storage

import (
	"io"
	"time"
)

type storageType string

const (
	S3Type    storageType = "s3"
	MinioType storageType = "minio"
	FakeType  storageType = "fake"
)

type Config struct {
	Type                storageType `envconfig:"type" default:"s3"`
	AwsKey              string      `envconfig:"aws_key"`
	AwsSecret           string      `envconfig:"aws_secret"`
	AwsRegion           string      `envconfig:"aws_region"`
	AwsBucket           string      `envconfig:"aws_bucket"`
	AwsEndpoint         string      `envconfig:"aws_endpoint" default:""`
	AwsDisableSSL       bool        `envconfig:"aws_disable_ssl" default:"false"`
	AwsS3ForcePathStyle bool        `envconfig:"aws_s3_force_path_style" default:"false"`
}

type Object struct {
	Name         string
	LastModified time.Time
}

type Storage interface {
	K8sSecretName() string
	AccessData() map[string][]byte
	UploadFile(path string, file io.ReadSeeker) error
	Type() string
	PodEnvVars() map[string]string
	List(path string) ([]*Object, error)
	Delete(path string) error
}

func New(conf *Config) (Storage, error) {
	switch conf.Type {
	case S3Type:
		return newS3(conf), nil
	case MinioType:
		return newMinio(conf), nil
	default:
		return nil, ErrInvalidStorageType
	}
}
