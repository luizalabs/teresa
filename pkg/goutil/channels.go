package goutil

import (
	"bufio"
	"io"
)

func LineGenerator(r io.Reader) (<-chan string, <-chan error) {
	ch := make(chan string)
	errCh := make(chan error)
	go func() {
		defer close(ch)
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			ch <- sc.Text()
		}
		if err := sc.Err(); err != nil {
			errCh <- err
		}
	}()
	return ch, errCh
}
