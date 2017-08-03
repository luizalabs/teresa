package storage

import (
	"fmt"
	"testing"
)

func TestMinioType(t *testing.T) {
	minio := newMinio(&Config{})
	if tmp := minio.Type(); tmp != "minio" {
		t.Errorf("expected minio, got %s", tmp)
	}
}

func TestMinioPodEnvVars(t *testing.T) {
	expectedHost := "url"
	expectedPort := "9000"
	expectedBucket := "test"
	minio := newMinio(&Config{
		AwsEndpoint: fmt.Sprintf("http://%s:%s", expectedHost, expectedPort),
		AwsBucket:   expectedBucket,
	})
	ev := minio.PodEnvVars()

	var testCases = []struct {
		env      string
		expected string
	}{
		{"S3_HOST", expectedHost},
		{"S3_PORT", expectedPort},
		{"MINIO_BUCKET", expectedBucket},
	}

	for _, tc := range testCases {
		if tmp := ev[tc.env]; tmp != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, tc.env)
		}
	}
}

func TestTrimEndpoint(t *testing.T) {
	var testCases = []struct {
		input    string
		expected string
	}{
		{"http://foo.com", "foo.com"},
		{"https://teresa.io", "teresa.io"},
	}

	for _, tc := range testCases {
		if actual := trimEndpoint(tc.input); actual != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, actual)
		}
	}
}
