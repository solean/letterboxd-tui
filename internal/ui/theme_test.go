package ui

import "testing"

func TestStarsToValue(t *testing.T) {
	if starsToValue("") != 0 {
		t.Fatalf("expected zero for empty rating")
	}
	if starsToValue("★★★½") != 3.5 {
		t.Fatalf("unexpected stars value")
	}
}

func TestGlowStars(t *testing.T) {
	out := glowStars("★★½")
	if stripANSI(out) != "★★½" {
		t.Fatalf("unexpected glowStars output: %q", stripANSI(out))
	}
}

func TestStyleRating(t *testing.T) {
	theme := newTheme()
	out := styleRating("★★★", theme)
	if stripANSI(out) != "★★★" {
		t.Fatalf("unexpected styleRating: %q", stripANSI(out))
	}
	if styleRating("", theme) != "" {
		t.Fatalf("expected empty output for empty rating")
	}
}
