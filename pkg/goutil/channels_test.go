package goutil

import (
	"strings"
	"testing"
	"testing/iotest"
)

func TestLineGeneratorOK(t *testing.T) {
	r := strings.NewReader("test")
	ch, _ := LineGenerator(r)

	if s := <-ch; s != "test" {
		t.Errorf("got %s; want test", s)
	}
}

func TestLineGeneratorError(t *testing.T) {
	r := iotest.TimeoutReader(strings.NewReader("test"))
	ch, errCh := LineGenerator(r)
	<-ch

	if err := <-errCh; err != iotest.ErrTimeout {
		t.Errorf("got %v; want %v", err, iotest.ErrTimeout)
	}
}
