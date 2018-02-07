package url

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFetchToTempOK(t *testing.T) {
	s := "fake://just-a-test"
	tmp, err := FetchToTemp(s)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp)

	file, err := os.Open(tmp)
	if err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadAll(file)
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != s {
		t.Errorf("got %s, want %s", b, s)
	}
}

func TestFetchToTempInvalidScheme(t *testing.T) {
	if _, err := FetchToTemp("test://just-a-test"); err == nil {
		t.Error("got nil, want error")
	}
}

func TestSchemeOK(t *testing.T) {
	var testCases = []struct {
		url string
		sch string
	}{
		{"http://a", "http"},
		{"://b", ""},
		{":/c", ""},
		{"test", ""},
	}

	for _, tc := range testCases {
		sch := Scheme(tc.url)
		if sch != tc.sch {
			t.Errorf("got %s, want %s", sch, tc.sch)
		}
	}
}
