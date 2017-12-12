package resource

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"testing"
)

type errReader struct {
	Err error
}

func (e *errReader) Read(p []byte) (int, error) {
	return 0, e.Err
}

func TestExecuteSuccess(t *testing.T) {
	exe := NewExecuter()
	r, err := os.Open("testdata/success.tmpl")
	if err != nil {
		t.Fatal("got unexpected error:", err)
	}
	var buf bytes.Buffer
	settings := []*Setting{
		&Setting{Key: "key1", Value: "value1"},
		&Setting{Key: "key2", Value: "value2"},
	}
	want := "value1 value2\n"

	if err := exe.Execute(&buf, r, settings); err != nil {
		t.Error("got error executing template:", err)
	}

	if buf.String() != want {
		t.Errorf("got %s, want %s", buf.String(), want)
	}
}

func TestExecuteErrReader(t *testing.T) {
	exe := NewExecuter()
	r := ioutil.NopCloser(&errReader{Err: errors.New("test")})
	var buf bytes.Buffer
	settings := []*Setting{
		&Setting{Key: "key1", Value: "value1"},
		&Setting{Key: "key2", Value: "value2"},
	}

	if err := exe.Execute(&buf, r, settings); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestExecuteFail(t *testing.T) {
	exe := NewExecuter()
	r, err := os.Open("testdata/fail.tmpl")
	if err != nil {
		t.Fatal("got unexpected error:", err)
	}
	var buf bytes.Buffer
	settings := []*Setting{
		&Setting{Key: "key1", Value: "value1"},
		&Setting{Key: "key2", Value: "value2"},
	}

	if err := exe.Execute(&buf, r, settings); err == nil {
		t.Error("expected error, got nil")
	}
}
