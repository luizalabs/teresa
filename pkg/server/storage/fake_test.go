package storage

import (
	"testing"
)

func TestFakeK8sSecretName(t *testing.T) {
	fake := NewFake()

	if name := fake.K8sSecretName(); name != "fake-storage" {
		t.Errorf("expected fake-storage, got %s", name)
	}
}

func TestFakeType(t *testing.T) {
	fake := NewFake()

	if tmp := fake.Type(); tmp != "fake" {
		t.Errorf("expected fake, got %s", tmp)
	}
}

func TestFakeAccessData(t *testing.T) {
	ad := NewFake().AccessData()
	var testCases = []struct {
		key   string
		value string
	}{
		{"region", "region"},
		{"builder-bucket", "bucket"},
		{"accesskey", "key"},
		{"secretkey", "secret"},
	}

	for _, tc := range testCases {
		got := string(ad[tc.key])
		if got != tc.value {
			t.Errorf("expected %s, got %s", tc.value, got)
		}
	}
}

func TestFakeUploadFile(t *testing.T) {
	fake := NewFake()

	if err := fake.UploadFile("/test", &fakeReadSeeker{}); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestFakePodEnvVars(t *testing.T) {
	fake := NewFake()
	ev := fake.PodEnvVars()
	if len(ev) != 0 {
		t.Errorf("expected 0, got %d", len(ev))
	}
}

func TestFakeList(t *testing.T) {
	fake := NewFake()
	items, err := fake.List("foo/bar")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if items[0].Name != "fake" || items[1].Name != "file" { // check fake.go
		t.Errorf("expected [fake file], got [%s %s]", items[0].Name, items[1].Name)
	}
}
