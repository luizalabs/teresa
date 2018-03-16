package app

import "testing"

func TestIsCronJob(t *testing.T) {
	var testCases = []struct {
		processType string
		expected    bool
	}{
		{"cron", true},
		{"cron-nro1", true},
		{"cronjob", true},
		{"web", false},
		{"worker", false},
		{"beat", false},
	}

	for _, tc := range testCases {
		if actual := IsCronJob(tc.processType); actual != tc.expected {
			t.Errorf("expected %v, got %v [case: %s]", tc.expected, actual, tc.processType)
		}
	}
}
