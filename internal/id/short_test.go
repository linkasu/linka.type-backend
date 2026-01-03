package id

import (
	"strings"
	"testing"
)

func TestNewShort(t *testing.T) {
	val := NewShort()
	if len(val) != shortSize {
		t.Fatalf("expected length %d, got %d", shortSize, len(val))
	}
	for _, r := range val {
		if !strings.ContainsRune(shortAlphabet, r) {
			t.Fatalf("unexpected character: %q", r)
		}
	}
}
