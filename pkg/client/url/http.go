package url

import (
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

type HTTPFetcher struct {
	cli HTTPClient
}

func (h *HTTPFetcher) Fetch(url string) (io.ReadCloser, error) {
	res, err := h.client().Get(url)
	if err != nil {
		return nil, errors.Wrap(err, "get failed")
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("got status code %d on url %s", res.StatusCode, url)
	}

	return res.Body, nil
}

func (h *HTTPFetcher) client() HTTPClient {
	if h.cli != nil {
		return h.cli
	}
	return http.DefaultClient
}
