package url

import (
	"io"
	"io/ioutil"
	"strings"
)

type FakeFetcher struct{}

func (f *FakeFetcher) Fetch(url string) (io.ReadCloser, error) {
	return ioutil.NopCloser(strings.NewReader(url)), nil
}
