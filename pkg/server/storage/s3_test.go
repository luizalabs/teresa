package storage

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
)

type fakeReadSeeker struct{}

type fakeS3Client struct{}

func (f *fakeReadSeeker) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (f *fakeReadSeeker) Read(p []byte) (int, error) {
	return 0, nil
}

func (f *fakeS3Client) PutObject(*s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return nil, nil
}

func TestS3GetK8sSecretName(t *testing.T) {
	s3 := newS3(&Config{})

	if name := s3.GetK8sSecretName(); name != "s3-storage" {
		t.Errorf("expected s3-storage, got %s", name)
	}
}

func TestS3Type(t *testing.T) {
	s3 := newS3(&Config{})

	if tmp := s3.Type(); tmp != "s3" {
		t.Errorf("expected s3, got %s", tmp)
	}
}

func TestS3GetAccessData(t *testing.T) {
	conf := &Config{
		AwsRegion: "region",
		AwsSecret: "secret",
		AwsKey:    "key",
		AwsBucket: "bucket",
	}
	s3 := newS3(conf)
	ad := s3.GetAccessData()

	if string(ad["region"]) != conf.AwsRegion {
		t.Errorf("expected %s, got %s", conf.AwsRegion, ad["region"])
	}
	if string(ad["builder-bucket"]) != conf.AwsBucket {
		t.Errorf("expected %s, got %s", conf.AwsBucket, ad["builder-bucket"])
	}
	if string(ad["accesskey"]) != conf.AwsKey {
		t.Errorf("expected %s, got %s", conf.AwsKey, ad["accesskey"])
	}
	if string(ad["secretkey"]) != conf.AwsSecret {
		t.Errorf("expected %s, got %s", conf.AwsSecret, ad["secretkey"])
	}
}

func TestS3UploadFile(t *testing.T) {
	s3 := newS3(&Config{})
	s3.(*S3).Client = &fakeS3Client{}

	if err := s3.UploadFile("/test", &fakeReadSeeker{}); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
