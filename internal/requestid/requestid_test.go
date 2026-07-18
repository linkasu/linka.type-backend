package requestid

import (
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	t.Run("keeps bounded safe value", func(t *testing.T) {
		const value = "client-Request_1.2"
		if got := Normalize(value); got != value {
			t.Fatalf("Normalize() = %q, want %q", got, value)
		}
	})

	for _, value := range []string{"", "unsafe/request", strings.Repeat("a", maxLength+1)} {
		t.Run(value, func(t *testing.T) {
			got := Normalize(value)
			if got == "" || got == value {
				t.Fatalf("Normalize(%q) = %q, want a generated request ID", value, got)
			}
		})
	}
}
