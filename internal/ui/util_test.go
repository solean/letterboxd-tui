package ui

import (
	"strings"
	"testing"
)

func TestMinMaxClamp(t *testing.T) {
	if max(1, 2) != 2 || min(1, 2) != 1 {
		t.Fatalf("unexpected min/max")
	}
	if clamp(5, 0, 3) != 3 || clamp(-1, 0, 3) != 0 {
		t.Fatalf("unexpected clamp")
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("hello", 0); got != "hello" {
		t.Fatalf("unexpected truncate: %q", got)
	}
	if got := stripANSI(truncate("hello", 3)); got == "hello" {
		t.Fatalf("expected truncation, got %q", got)
	}
}

func TestCompactSpaces(t *testing.T) {
	if got := compactSpaces("  a   b  "); got != "a b" {
		t.Fatalf("unexpected compactSpaces: %q", got)
	}
}

func TestWrapText(t *testing.T) {
	text := "one two three"
	if got := wrapText(text, 0); got != text {
		t.Fatalf("unexpected wrapText for width 0: %q", got)
	}
	if got := wrapText(text, 6); got == text {
		t.Fatalf("expected wrapped text, got %q", got)
	}
}

func TestAppendWithSpacing(t *testing.T) {
	var b strings.Builder
	appendWithSpacing(&b, "a")
	appendWithSpacing(&b, "b")
	if got := b.String(); got != "a b" {
		t.Fatalf("unexpected builder: %q", got)
	}
}

func TestModalDimensions(t *testing.T) {
	w, h := modalDimensions(120, 40)
	if w <= 0 || h <= 0 {
		t.Fatalf("unexpected dimensions: %d %d", w, h)
	}
	if got := modalContentWidth(120, 40); got <= 0 {
		t.Fatalf("unexpected content width: %d", got)
	}
}

func TestFormatWhen(t *testing.T) {
	if got := formatWhen(""); got != "" {
		t.Fatalf("expected empty format")
	}
	if got := formatWhen("not-a-time"); got != "not-a-time" {
		t.Fatalf("unexpected format for invalid time: %q", got)
	}
	if got := formatWhen("2024-01-05T00:00:00Z"); got != "Jan 05 2024" {
		t.Fatalf("unexpected format: %q", got)
	}
}
