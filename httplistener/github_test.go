package httplistener

import (
	"testing"
)

func TestIsMergeQueueBranch(t *testing.T) {
	tests := []struct {
		ref      string
		expected bool
	}{
		{"refs/heads/master", false},
		{"refs/heads/gh-readonly-queue/master/pr-1", true},
		{"gh-readonly-queue/master/pr-1", true},
		{"main", false},
		{"refs/tags/v1.0.0", false},
	}

	for _, tc := range tests {
		if got := isMergeQueueBranch(tc.ref); got != tc.expected {
			t.Errorf("isMergeQueueBranch(%q) = %v; want %v", tc.ref, got, tc.expected)
		}
	}
}
