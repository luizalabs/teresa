package server

import (
	"bytes"
	"fmt"
	"testing"
)

func TestFilteredWriter(t *testing.T) {
	var testCases = []struct {
		filter   [][]byte
		in       string
		expected string
	}{
		{[][]byte{[]byte("please ignore")}, "useless msg please ignore", ""},
		{[][]byte{}, "a message", "a message"},
		{[][]byte{[]byte("please ignore")}, "very important info", "very important info"},
		{[][]byte{[]byte("please ignore"), []byte("useless")}, "useless msg", ""},
	}

	for _, tc := range testCases {
		msgsToFilter = tc.filter

		b := new(bytes.Buffer)
		w := &FilteredWriter{Writer: b}

		fmt.Fprint(w, tc.in)

		actual := b.String()
		if actual != tc.expected {
			t.Errorf("expected %s, got %s", tc.expected, actual)
		}
	}
}
