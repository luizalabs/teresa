package resource

import (
	"fmt"
	"io"
	"net/http"
)

type Templater interface {
	Template(resName string) (io.ReadCloser, error)
	WelcomeTemplate(resName string) (io.ReadCloser, error)
}

type HTTPClient interface {
	Get(url string) (res *http.Response, err error)
}

type httpTemplater struct {
	urlFmt        string
	welcomeURLFmt string
	cli           HTTPClient
}

func (h *httpTemplater) get(myFmt, resName string) (io.ReadCloser, error) {
	url := fmt.Sprintf(myFmt, resName)

	res, err := h.cli.Get(url)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got status code %d on url %s", res.StatusCode, url)
	}

	return res.Body, nil
}

func (h *httpTemplater) Template(resName string) (io.ReadCloser, error) {
	return h.get(h.urlFmt, resName)
}

func (h *httpTemplater) WelcomeTemplate(resName string) (io.ReadCloser, error) {
	return h.get(h.welcomeURLFmt, resName)
}

func NewTemplater(cfg *Config, cli HTTPClient) Templater {
	return &httpTemplater{
		urlFmt:        cfg.URLFmt,
		welcomeURLFmt: cfg.WelcomeURLFmt,
		cli:           cli,
	}
}
