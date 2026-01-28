package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"

	"github.com/solean/letterboxd-tui/internal/letterboxd"
)

type logForm struct {
	active       bool
	focus        int
	rating       textinput.Model
	date         textinput.Model
	tags         textinput.Model
	review       textarea.Model
	rewatch      bool
	spoilers     bool
	liked        bool
	draft        bool
	privacyIndex int
	status       string
	submitting   bool
}

const (
	logFieldRating = iota
	logFieldRewatch
	logFieldDate
	logFieldReview
	logFieldSpoilers
	logFieldLiked
	logFieldTags
	logFieldPrivacy
	logFieldDraft
	logFieldSubmit
	logFieldCount
)

var privacyOptions = []string{"Default", "Anyone", "Friends", "You"}

func newLogForm(film letterboxd.Film) logForm {
	rating := textinput.New()
	rating.Placeholder = "e.g. 4.5"
	rating.CharLimit = 4
	if film.UserRating != "" {
		val := starsToValue(film.UserRating) / 2.0
		if val > 0 {
			rating.SetValue(strings.TrimRight(strings.TrimRight(fmtFloat(val, 1), "0"), "."))
		}
	}

	date := textinput.New()
	date.Placeholder = "YYYY-MM-DD"
	date.CharLimit = 10
	date.SetValue(time.Now().Format("2006-01-02"))

	tags := textinput.New()
	tags.Placeholder = "comma,separated"
	tags.CharLimit = 200

	review := textarea.New()
	review.SetHeight(5)
	review.SetWidth(48)
	review.Placeholder = "Add a review..."

	return logForm{
		active: true,
		focus:  logFieldRating,
		rating: rating,
		date:   date,
		tags:   tags,
		review: review,
	}
}

func (f *logForm) setSize(width int) {
	target := max(30, min(80, width-10))
	f.rating.Width = min(12, target)
	f.date.Width = min(12, target)
	f.tags.Width = target
	f.review.SetWidth(target)
}

func (f *logForm) focusField(idx int) {
	if idx < 0 {
		idx = 0
	}
	if idx >= logFieldCount {
		idx = logFieldCount - 1
	}
	f.focus = idx
	f.rating.Blur()
	f.date.Blur()
	f.tags.Blur()
	f.review.Blur()
	switch idx {
	case logFieldRating:
		f.rating.Focus()
	case logFieldDate:
		f.date.Focus()
	case logFieldTags:
		f.tags.Focus()
	case logFieldReview:
		f.review.Focus()
	}
}

func (f logForm) privacyLabel() string {
	idx := clamp(f.privacyIndex, 0, len(privacyOptions)-1)
	return privacyOptions[idx]
}

func fmtFloat(val float64, precision int) string {
	format := "%." + strconv.Itoa(precision) + "f"
	out := fmt.Sprintf(format, val)
	out = strings.TrimRight(out, "0")
	out = strings.TrimRight(out, ".")
	return out
}
