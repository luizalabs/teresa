package cmd

import "testing"

func TestParseURL(t *testing.T) {
	goodUrls := []string{
		"http://127.0.0.1:8080",
		"http://4.2.2.2",
		"https://myserver.com",
	}
	badUrls := []string{
		"127.0.0.1:8080",
		"foobar",
		"4.2.2.2",
		"myserver.com",
	}
	for i := range goodUrls {
		_, err := ParseServerURL(goodUrls[i])
		if err != nil {
			t.Errorf("Parsing should have passed for url (%s), error: %+v", goodUrls[i], err)
		}
	}
	for i := range goodUrls {
		_, err := ParseServerURL(badUrls[i])
		if err == nil {
			t.Errorf("Parsing should have failed for url (%s)", badUrls[i])
		}
	}
}
