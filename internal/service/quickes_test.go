package service

import (
	"testing"

	"github.com/linkasu/linka.type-backend/internal/defaults"
)

func TestNormalizeQuickes(t *testing.T) {
	input := []string{"Привет", "", "Да"}
	got := normalizeQuickes(input)
	if len(got) != len(defaults.DefaultQuickes) {
		t.Fatalf("expected %d quickes, got %d", len(defaults.DefaultQuickes), len(got))
	}
	if got[0] != "Привет" {
		t.Fatalf("expected custom quickes preserved")
	}
	if got[1] != defaults.DefaultQuickes[1] {
		t.Fatalf("expected default fallback for empty slot")
	}
	if got[2] != "Да" {
		t.Fatalf("expected custom quickes preserved for slot 2")
	}
}
