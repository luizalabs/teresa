package helpers

import (
	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-openapi/runtime"
	"github.com/kelseyhightower/envconfig"
)

type storageType string

const (
	awsS3 storageType = "s3"
)

type storageConfig struct {
	Type      storageType `envconfig:"type"`
	AwsKey    string      `envconfig:"aws_key"`
	AwsSecret string      `envconfig:"aws_secret"`
	AwsRegion string      `envconfig:"aws_region"`
	AwsBucket string      `envconfig:"aws_bucket"`
}

// Storage is an interface used to accept more than one type of file storage
type Storage interface {
	GetK8sSecretName() string
	GetAccessData() map[string][]byte
	UploadFile(path string, file *runtime.File) error
	Type() string
}

type awsS3Storage struct {
	Client *s3.S3
	Key    string
	Secret string
	Region string
	Bucket string
}

func (a *awsS3Storage) GetK8sSecretName() string {
	return "s3-storage"
}

func (a *awsS3Storage) GetAccessData() map[string][]byte {
	return map[string][]byte{
		"region":         []byte(a.Region),
		"builder-bucket": []byte(a.Bucket),
		"accesskey":      []byte(a.Key),
		"secretkey":      []byte(a.Secret),
	}
}

func (a *awsS3Storage) UploadFile(path string, file *runtime.File) error {
	defer file.Data.Close()
	po := &s3.PutObjectInput{
		Bucket: &a.Bucket,
		Body:   file.Data,
		Key:    &path,
	}
	if _, err := a.Client.PutObject(po); err != nil {
		return err
	}
	return nil
}

func (a *awsS3Storage) Type() string {
	return string(awsS3)
}

// FileStorage is used to concentrate all storage functions in only one place
// This is a helper to abstract the providers
var FileStorage Storage

func init() {
	conf := &storageConfig{}
	err := envconfig.Process("teresafilestorage", conf)

	if err != nil {
		log.Fatalf("failed to read the storage configuration from environment: %s", err.Error())
	}

	if conf.Type == awsS3 {
		if conf.AwsKey == "" || conf.AwsSecret == "" || conf.AwsRegion == "" || conf.AwsBucket == "" {
			log.Fatalf("failed to read s3 configuration from environment: %s", err.Error())
		}
		st := &awsS3Storage{
			Key:    conf.AwsKey,
			Region: conf.AwsRegion,
			Secret: conf.AwsSecret,
			Bucket: conf.AwsBucket,
		}

		st.Client = s3.New(session.New(), &aws.Config{
			Region:      &st.Region,
			Credentials: credentials.NewStaticCredentials(st.Key, st.Secret, ""),
		})

		FileStorage = st

	} else {
		log.Fatalf("no file storage specified")
	}
}
