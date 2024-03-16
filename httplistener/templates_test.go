package httplistener

import "testing"

func TestTemplates(t *testing.T) {
	test_refName(t, "refs/heads/main", "main")
	test_refName(t, "refs/tags/v1.2.3", "v1.2.3")
	test_refName(t, "refs/heads/feature/123-abc", "feature/123-abc")
}

func test_refName(t *testing.T, ref string, expected string) {
	got := refName(ref)
	if got != expected {
		t.Fatalf("Expected ref %q to be prettified to %q, got %q", ref, expected, got)
	}
}
