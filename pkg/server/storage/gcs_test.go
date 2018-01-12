package storage

import (
	"io"
	"testing"
)

type fakeGcsClient struct{}

func (f *fakeGcsClient) PutObject(bucket string, path string, file io.ReadSeeker) error {
	return nil
}

func TestGcsK8sSecretName(t *testing.T) {
	gcs := &Gcs{Client: &fakeGcsClient{}}
	if name := gcs.K8sSecretName(); name != "gcs-storage" {
		t.Errorf("expected gcs-storage, got %s", name)
	}
}

func TestGcsType(t *testing.T) {
	gcs := &Gcs{Client: &fakeGcsClient{}}
	if tmp := gcs.Type(); tmp != "gcs" {
		t.Errorf("expected gcs, got %s", tmp)
	}
}

func TestGcsAccessData(t *testing.T) {
	expectedBucket := "gcsBucket"
	expectedKeyJson := "{foo:bar}"

	gcs := &Gcs{Bucket: expectedBucket, Cred: expectedKeyJson, Client: &fakeGcsClient{}}
	ad := gcs.AccessData()
	if actual := string(ad["builder-bucket"]); actual != expectedBucket {
		t.Errorf("expected %s, got %s", expectedBucket, actual)
	}
	if actual := string(ad["key.json"]); actual != expectedKeyJson {
		t.Errorf("expected %s, got %s", expectedKeyJson, actual)
	}
}

func TestGcsUploadFile(t *testing.T) {
	gcs := &Gcs{Client: &fakeGcsClient{}}
	if err := gcs.UploadFile("/path", &fakeReadSeeker{}); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestGcsPodEnvVars(t *testing.T) {
	gcs := &Gcs{}
	ev := gcs.PodEnvVars()
	if len(ev) != 0 {
		t.Errorf("expected 0, got %d", len(ev))
	}
}
