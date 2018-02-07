package url

import (
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
)

type Fetcher interface {
	Fetch(url string) (io.ReadCloser, error)
}

func FetchToTemp(url string) (string, error) {
	fet, err := newFetcher(url)
	if err != nil {
		return "", err
	}

	rc, err := fet.Fetch(url)
	if err != nil {
		return "", errors.Wrap(err, "fetch failed")
	}
	defer rc.Close()

	tmp, err := ioutil.TempFile("", "teresa")
	if err != nil {
		return "", errors.Wrap(err, "failed to create temp file")
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, rc); err != nil {
		os.Remove(tmp.Name())
		return "", errors.Wrap(err, "copy failed")
	}

	return tmp.Name(), nil
}

func Scheme(url string) string {
	idx := strings.Index(url, "://")
	if idx <= 0 {
		return ""
	}
	return url[:idx]
}

func newFetcher(url string) (Fetcher, error) {
	sch := Scheme(url)
	switch sch {
	case "http", "https":
		return &HTTPFetcher{}, nil
	case "fake":
		return &FakeFetcher{}, nil
	default:
		return nil, errors.New("invalid scheme")
	}
}
