package templates

import "testing"

func TestGetAndList(t *testing.T) {
	names := List()
	if len(names) < 4 {
		t.Fatalf("expected built-in templates")
	}
	if _, err := Get("tervux"); err != nil {
		t.Fatalf("expected tervux template: %v", err)
	}
	if _, err := Get("missing"); err == nil {
		t.Fatalf("expected error for unknown template")
	}
}
