package storage

import (
	"io"
	"time"
)

type fake struct {
	Key    string
	Secret string
	Region string
	Bucket string
}

func (f *fake) K8sSecretName() string {
	return "fake-storage"
}

func (f *fake) AccessData() map[string][]byte {
	return map[string][]byte{
		"region":         []byte(f.Region),
		"builder-bucket": []byte(f.Bucket),
		"accesskey":      []byte(f.Key),
		"secretkey":      []byte(f.Secret),
	}
}

func (f *fake) UploadFile(path string, file io.ReadSeeker) error {
	return nil
}

func (f *fake) List(path string) ([]*Object, error) {
	return []*Object{
		&Object{Name: "fake", LastModified: time.Now()},
		&Object{Name: "file", LastModified: time.Now()},
	}, nil
}

func (f *fake) Delete(path string) error {
	return nil
}

func (f *fake) Type() string {
	return string(FakeType)
}

func (f *fake) PodEnvVars() map[string]string {
	return make(map[string]string)
}

func NewFake() *fake {
	return &fake{
		Key:    "key",
		Region: "region",
		Secret: "secret",
		Bucket: "bucket",
	}
}
