package app

import "testing"

func TestSortEnvsByKey(t *testing.T) {
	items := []*EnvVar{
		{
			Key:   "B",
			Value: "1",
		},
		{
			Key:   "A",
			Value: "2",
		},
	}

	sortEnvsByKey(items)

	if items[0].Key != "A" {
		t.Errorf("expected A, got %s", items[0].Key)
	}
}
