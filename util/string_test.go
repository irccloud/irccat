package util

import (
	"testing"
)

func TestTruncate(t *testing.T) {
	if Truncate("Test", 5) != "Test" {
		t.Fail()
	}
	if Truncate("Test", 3) != "" {
		t.Fail()
	}
	if Truncate("This is a test", 9) != "This is…" {
		t.Fail()
	}
	if Truncate("This is a test", 10) != "This is a…" {
		t.Fail()
	}

}
