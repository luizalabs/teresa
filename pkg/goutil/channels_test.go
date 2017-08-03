package goutil

import (
	"bytes"
	"testing"
)

func TestChannelFromReader(t *testing.T) {
	var testCases = []struct {
		ln       bool
		expected string
	}{
		{true, "test\n"},
		{false, "test"},
	}

	for _, tc := range testCases {
		buf := bytes.NewBufferString(tc.expected)
		ch := ChannelFromReader(buf, tc.ln)

		s := <-ch
		if s != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, s)
		}
	}
}
