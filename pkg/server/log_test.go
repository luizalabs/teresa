package server

import (
	"bytes"
	"fmt"
	"testing"
)

func TestFilteredWriter(t *testing.T) {
	var testCases = []struct {
		filter   []string
		in       string
		expected string
	}{
		{[]string{"please ignore"}, "useless msg please ignore", ""},
		{[]string{}, "a message", "a message"},
		{[]string{"please ignore"}, "very important info", "very important info"},
		{[]string{"please ignore", "useless"}, "useless msg", ""},
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
