package module

import "testing"

func TestGMMoveToText(t *testing.T) {
	got := GMMoveToText(1, 77, 73)
	want := "@MoveTo 1 77 73"
	if got != want {
		t.Fatalf("GMMoveToText() = %q, want %q", got, want)
	}
}
