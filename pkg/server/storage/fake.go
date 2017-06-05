package storage

import (
	"io"
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

func (f *fake) Type() string {
	return string(FakeType)
}

func NewFake() Storage {
	return &fake{
		Key:    "key",
		Region: "region",
		Secret: "secret",
		Bucket: "bucket",
	}
}
