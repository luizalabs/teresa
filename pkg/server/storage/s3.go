package storage

import (
	"io"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type S3Client interface {
	PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
	ListObjects(*s3.ListObjectsInput) (*s3.ListObjectsOutput, error)
}

type S3 struct {
	Client           S3Client
	Key              string
	Secret           string
	Region           string
	Bucket           string
	Endpoint         string
	DisableSSL       bool
	S3ForcePathStyle bool
}

func (s *S3) K8sSecretName() string {
	return "s3-storage"
}

func (s *S3) AccessData() map[string][]byte {
	return map[string][]byte{
		"region":         []byte(s.Region),
		"builder-bucket": []byte(s.Bucket),
		"accesskey":      []byte(s.Key),
		"secretkey":      []byte(s.Secret),
	}
}

func (s *S3) UploadFile(path string, file io.ReadSeeker) error {
	po := &s3.PutObjectInput{
		Bucket: &s.Bucket,
		Body:   file,
		Key:    &path,
	}
	_, err := s.Client.PutObject(po)
	return err
}

func (s *S3) List(path string) ([]*Object, error) {
	li := &s3.ListObjectsInput{
		Bucket: &s.Bucket,
		Prefix: aws.String(path),
	}

	res, err := s.Client.ListObjects(li)
	if err != nil {
		return nil, err
	}

	out := []*Object{}
	m := make(map[string]bool)
	for _, item := range res.Contents {
		name := strings.TrimPrefix(*item.Key, path)
		name = strings.Split(name, "/")[0]
		if _, found := m[name]; !found {
			m[name] = true
			out = append(out, &Object{Name: name, LastModified: *item.LastModified})
		}
	}

	return out, nil
}

func (s *S3) Type() string {
	return string(S3Type)
}

func (s *S3) PodEnvVars() map[string]string {
	return make(map[string]string)
}

func newS3(conf *Config) *S3 {
	st := &S3{
		Key:              conf.AwsKey,
		Region:           conf.AwsRegion,
		Secret:           conf.AwsSecret,
		Bucket:           conf.AwsBucket,
		Endpoint:         conf.AwsEndpoint,
		DisableSSL:       conf.AwsDisableSSL,
		S3ForcePathStyle: conf.AwsS3ForcePathStyle,
	}
	awsConf := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(st.Key, st.Secret, ""),
		Region:           &st.Region,
		Endpoint:         &st.Endpoint,
		DisableSSL:       &st.DisableSSL,
		S3ForcePathStyle: &st.S3ForcePathStyle,
	}
	st.Client = s3.New(session.New(), awsConf)
	return st
}
