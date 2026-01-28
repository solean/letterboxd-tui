package ui

import (
	"testing"
	"time"

	"github.com/solean/letterboxd-tui/internal/letterboxd"
)

func TestNewLogForm(t *testing.T) {
	film := letterboxd.Film{UserRating: "★★★½"}
	form := newLogForm(film)
	if form.rating.Value() == "" {
		t.Fatalf("expected rating to be prefilled")
	}
	if _, err := time.Parse("2006-01-02", form.date.Value()); err != nil {
		t.Fatalf("expected date format, got %q", form.date.Value())
	}
}

func TestLogFormSetSize(t *testing.T) {
	form := newLogForm(letterboxd.Film{})
	form.setSize(120)
	if form.tags.Width == 0 || form.review.Width() == 0 {
		t.Fatalf("expected sizes to be set")
	}
}

func TestLogFormFocusField(t *testing.T) {
	form := newLogForm(letterboxd.Film{})
	form.focusField(logFieldTags)
	if form.focus != logFieldTags {
		t.Fatalf("unexpected focus: %d", form.focus)
	}
	form.focusField(-1)
	if form.focus != logFieldRating {
		t.Fatalf("expected clamp to first field, got %d", form.focus)
	}
	form.focusField(logFieldCount + 1)
	if form.focus != logFieldSubmit {
		t.Fatalf("expected clamp to last field, got %d", form.focus)
	}
}

func TestPrivacyLabel(t *testing.T) {
	form := newLogForm(letterboxd.Film{})
	form.privacyIndex = 2
	if form.privacyLabel() != "Friends" {
		t.Fatalf("unexpected privacy label")
	}
}

func TestFmtFloat(t *testing.T) {
	if got := fmtFloat(1.50, 1); got != "1.5" {
		t.Fatalf("unexpected fmtFloat: %q", got)
	}
	if got := fmtFloat(1.00, 1); got != "1" {
		t.Fatalf("unexpected fmtFloat: %q", got)
	}
}
