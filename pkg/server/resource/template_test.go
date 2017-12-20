package resource

import (
	"errors"
	"net/http"
	"testing"
)

type fakeHTTPClient struct {
	Err        error
	StatusCode int
}

func (f *fakeHTTPClient) Get(url string) (*http.Response, error) {
	res := &http.Response{Body: &fakeReadCloser{}, StatusCode: f.StatusCode}
	return res, f.Err
}

func TestTemplateSuccess(t *testing.T) {
	cfg := &Config{}
	tpl := NewTemplater(cfg, &fakeHTTPClient{StatusCode: http.StatusOK})

	if _, err := tpl.Template("test"); err != nil {
		t.Error("got error fetching template:", err)
	}
}

func TestTemplateErrHTTPProtocol(t *testing.T) {
	cfg := &Config{}
	tpl := NewTemplater(cfg, &fakeHTTPClient{Err: errors.New("test")})

	if _, err := tpl.Template("test"); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestTemplateErrStatusCode(t *testing.T) {
	cfg := &Config{URLFmt: "http://%s"}
	tpl := NewTemplater(cfg, &fakeHTTPClient{})

	if _, err := tpl.Template("test"); err == nil {
		t.Error("expected error, got nil")
	}
}
