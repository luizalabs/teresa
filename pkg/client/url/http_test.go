package url

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
)

type fakeHTTPClient struct {
	err error
	sc  int
}

func (f *fakeHTTPClient) Get(url string) (*http.Response, error) {
	if f.err != nil {
		return nil, f.err
	}

	sc := f.sc
	if sc == 0 {
		sc = http.StatusOK
	}

	return &http.Response{
		Body:       ioutil.NopCloser(strings.NewReader("test")),
		StatusCode: sc,
	}, nil
}

func TestHTTPFetcherOK(t *testing.T) {
	fet := HTTPFetcher{&fakeHTTPClient{}}

	if _, err := fet.Fetch("test"); err != nil {
		t.Error("want nil; got error", err)
	}
}

func TestHTTPFetcherFail(t *testing.T) {
	fet := HTTPFetcher{&fakeHTTPClient{err: errors.New("test")}}

	if _, err := fet.Fetch("test"); err == nil {
		t.Error("want error; got nil")
	}
}

func TestHTTPFetcherFailStatusCode(t *testing.T) {
	fet := HTTPFetcher{&fakeHTTPClient{sc: http.StatusNotFound}}

	if _, err := fet.Fetch("test"); err == nil {
		t.Error("want error; got nil")
	}
}
