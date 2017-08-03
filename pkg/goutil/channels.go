package goutil

import (
	"bufio"
	"fmt"
	"io"
)

func ChannelFromReader(r io.Reader, ln bool) <-chan string {
	fn := fmt.Sprint
	if ln {
		fn = fmt.Sprintln
	}

	c := make(chan string)
	go func() {
		defer close(c)
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			c <- fn(scanner.Text())
		}
	}()

	return c
}
