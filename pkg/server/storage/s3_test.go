package storage

import (
	"reflect"
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

func TestS3K8sSecretName(t *testing.T) {
	s3 := newS3(&Config{})

	if name := s3.K8sSecretName(); name != "s3-storage" {
		t.Errorf("expected s3-storage, got %s", name)
	}
}

func TestS3Type(t *testing.T) {
	s3 := newS3(&Config{})

	if tmp := s3.Type(); tmp != "s3" {
		t.Errorf("expected s3, got %s", tmp)
	}
}

func TestS3AccessData(t *testing.T) {
	conf := &Config{
		AwsRegion: "region",
		AwsSecret: "secret",
		AwsKey:    "key",
		AwsBucket: "bucket",
	}
	s3 := newS3(conf)
	ad := s3.AccessData()
	var testCases = []struct {
		key   string
		field string
	}{
		{"region", "AwsRegion"},
		{"builder-bucket", "AwsBucket"},
		{"accesskey", "AwsKey"},
		{"secretkey", "AwsSecret"},
	}

	for _, tc := range testCases {
		v := reflect.ValueOf(conf)
		expected := reflect.Indirect(v).FieldByName(tc.field).String()
		got := string(ad[tc.key])
		if got != expected {
			t.Errorf("expected %s, got %s", expected, got)
		}
	}
}

func TestS3UploadFile(t *testing.T) {
	s3 := newS3(&Config{})
	s3.(*S3).Client = &fakeS3Client{}

	if err := s3.UploadFile("/test", &fakeReadSeeker{}); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestS3PodEnvVars(t *testing.T) {
	s3 := newS3(&Config{})
	ev := s3.PodEnvVars()
	if len(ev) != 0 {
		t.Errorf("expected 0, got %d", len(ev))
	}
}
