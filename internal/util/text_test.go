package util

import (
	"testing"
)

func TestEnsureSpaceAfterDot(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Das ist ein Satz.Hier ist noch einer.", "Das ist ein Satz. Hier ist noch einer."},
		{"Das ist ein Satz. Hier ist noch einer.", "Das ist ein Satz. Hier ist noch einer."},
		{"Version 1.2 ist aktuell.", "Version 1.2 ist aktuell."},
		{"Ein Punkt am Ende.", "Ein Punkt am Ende."},
		{"Abkürzung z.B. Test.", "Abkürzung z. B. Test."},
		{"Mehrere Punkte...Test.", "Mehrere Punkte... Test."},
		{"Klammern (z.B.) bleiben.", "Klammern (z. B.) bleiben."},
		{"Das ist toll.Sollte man machen.", "Das ist toll. Sollte man machen."},
	}

	for _, tc := range tests {
		got := EnsureSpaceAfterDot(tc.input)
		if got != tc.expected {
			t.Errorf("EnsureSpaceAfterDot(%q) = %q; want %q", tc.input, got, tc.expected)
		}
	}
}
