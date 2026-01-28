package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"

	"github.com/solean/letterboxd-tui/internal/letterboxd"
)

func TestTabNavigation(t *testing.T) {
	withCookie := Model{client: &letterboxd.Client{Cookie: "foo=bar"}}
	if len(visibleTabs(withCookie)) == 0 {
		t.Fatalf("expected visible tabs")
	}
	if visibleTabByIndex(withCookie, -1) != tabProfile {
		t.Fatalf("expected profile for invalid index")
	}
	if nextTab(withCookie, tabSearch) != tabProfile {
		t.Fatalf("expected wrap to profile")
	}
	if prevTab(withCookie, tabProfile) != tabSearch {
		t.Fatalf("expected wrap to search")
	}
	for _, tab := range visibleTabs(Model{}) {
		if tab == tabFollowing {
			t.Fatalf("expected friends tab hidden without cookie")
		}
	}
}

func TestRenderTabs(t *testing.T) {
	m := Model{activeTab: tabDiary}
	out := stripANSI(renderTabs(m, newTheme()))
	if !strings.Contains(out, "Diary") || !strings.Contains(out, "Profile") {
		t.Fatalf("unexpected tabs output: %q", out)
	}
}

func TestRenderHelpVariants(t *testing.T) {
	theme := newTheme()
	m := NewModel("jane", nil)
	m.activeTab = tabSearch
	m.searchFocusInput = true
	if out := renderHelp(m, theme, 120); !strings.Contains(out, "search") {
		t.Fatalf("unexpected search help: %q", out)
	}
	m.searchFocusInput = false
	if out := renderHelp(m, theme, 120); !strings.Contains(out, "/") {
		t.Fatalf("unexpected search help: %q", out)
	}
	m = NewModel("jane", nil)
	m.activeTab = tabFilm
	m.film = letterboxd.Film{URL: letterboxd.BaseURL + "/film/inception/"}
	m.watchlistLoaded = true
	m.client = &letterboxd.Client{Cookie: "foo=bar"}
	if out := renderHelp(m, theme, 120); !strings.Contains(out, "add to watchlist") {
		t.Fatalf("unexpected film help: %q", out)
	}
	m.film.WatchlistOK = true
	m.film.InWatchlist = true
	if out := renderHelp(m, theme, 120); !strings.Contains(out, "remove from watchlist") {
		t.Fatalf("unexpected film help: %q", out)
	}
	m.client = nil
	if out := renderHelp(m, theme, 120); strings.Contains(out, "watchlist") || strings.Contains(out, "log entry") {
		t.Fatalf("expected cookie-required actions hidden: %q", out)
	}
	m = NewModel("jane", nil)
	m.profileModal = true
	if out := renderHelp(m, theme, 120); !strings.Contains(out, "back") {
		t.Fatalf("unexpected profile help: %q", out)
	}
}

func TestRenderDiary(t *testing.T) {
	theme := newTheme()
	m := Model{diaryErr: errDummy{}}
	if out := stripANSI(renderDiary(m, theme)); !strings.Contains(out, "Error:") {
		t.Fatalf("expected error output, got %q", out)
	}
	m = Model{loading: true}
	if out := stripANSI(renderDiary(m, theme)); !strings.Contains(out, "Loading diary") {
		t.Fatalf("expected loading output, got %q", out)
	}
	m = Model{diary: []letterboxd.DiaryEntry{{Title: "Inception", Rating: "★★★★", Rewatch: true, Review: true, Date: "Jan 5 2024"}}}
	out := stripANSI(renderDiary(m, theme))
	if !strings.Contains(out, "Inception") || !strings.Contains(out, "★★★★") {
		t.Fatalf("unexpected diary output: %q", out)
	}
	if !strings.Contains(out, "↺") || !strings.Contains(out, "✎") {
		t.Fatalf("expected flags in output: %q", out)
	}
}

func TestRenderWatchlist(t *testing.T) {
	theme := newTheme()
	m := Model{watchErr: errDummy{}}
	if out := stripANSI(renderWatchlist(m, theme)); !strings.Contains(out, "Error:") {
		t.Fatalf("expected error output")
	}
	m = Model{watchlist: []letterboxd.WatchlistItem{{Title: "Inception", Year: "2010"}}}
	out := stripANSI(renderWatchlist(m, theme))
	if !strings.Contains(out, "Inception (2010)") {
		t.Fatalf("unexpected watchlist output: %q", out)
	}
}

func TestRenderProfileContent(t *testing.T) {
	theme := newTheme()
	profile := letterboxd.Profile{
		Stats:     []letterboxd.ProfileStat{{Label: "Films", Value: "42"}},
		Favorites: []letterboxd.FavoriteFilm{{Title: "Inception", Year: "2010"}},
		Recent:    []string{"jane watched Inception"},
	}
	out := stripANSI(renderProfileContent(profile, nil, false, "jane", nil, theme))
	if !strings.Contains(out, "Films") || !strings.Contains(out, "Inception") {
		t.Fatalf("unexpected profile output: %q", out)
	}
}

func TestRenderSearch(t *testing.T) {
	theme := newTheme()
	m := Model{searchLoading: true}
	out := stripANSI(renderSearch(m, theme))
	if !strings.Contains(out, "Searching") {
		t.Fatalf("unexpected search output: %q", out)
	}
	m = Model{searchResults: []letterboxd.SearchResult{{Title: "Memento", Year: "2000"}}}
	out = stripANSI(renderSearch(m, theme))
	if !strings.Contains(out, "Memento (2000)") {
		t.Fatalf("unexpected search results: %q", out)
	}
}

func TestRenderFilm(t *testing.T) {
	theme := newTheme()
	m := Model{filmErr: errDummy{}}
	if out := stripANSI(renderFilm(m, theme)); !strings.Contains(out, "Error:") {
		t.Fatalf("expected error output")
	}
	m = Model{film: letterboxd.Film{Title: "Inception"}}
	out := stripANSI(renderFilm(m, theme))
	if !strings.Contains(out, "Inception") {
		t.Fatalf("unexpected film output: %q", out)
	}
}

func TestRenderLogForm(t *testing.T) {
	theme := newTheme()
	m := Model{film: letterboxd.Film{Title: "Inception"}, logForm: newLogForm(letterboxd.Film{})}
	out := stripANSI(renderLogForm(m, theme))
	if !strings.Contains(out, "Log diary entry") {
		t.Fatalf("unexpected log form output: %q", out)
	}
	if !strings.Contains(out, "Rewatch:") {
		t.Fatalf("unexpected log form output: %q", out)
	}
}

func TestRenderLogStatus(t *testing.T) {
	theme := newTheme()
	m := Model{logForm: logForm{submitting: true}, logSpinner: spinner.Model{}}
	out := stripANSI(renderLogStatus(m, theme))
	if !strings.Contains(out, "Submitting") {
		t.Fatalf("unexpected log status: %q", out)
	}
	m.logForm.submitting = false
	m.logForm.status = "Error: nope"
	out = stripANSI(renderLogStatus(m, theme))
	if !strings.Contains(out, "Error") {
		t.Fatalf("unexpected log status: %q", out)
	}
}

func TestRenderLogToggle(t *testing.T) {
	out := stripANSI(renderLogToggle("Like", true, false, newTheme()))
	if !strings.Contains(out, "Like: yes") {
		t.Fatalf("unexpected toggle output: %q", out)
	}
}

func TestRenderSelectableLine(t *testing.T) {
	out := stripANSI(renderSelectableLine("Line", true, 50, newTheme()))
	if !strings.Contains(out, "> Line") {
		t.Fatalf("unexpected selectable output: %q", out)
	}
}

func TestRenderReviews(t *testing.T) {
	theme := newTheme()
	out := stripANSI(renderReviews("Title", nil, errDummy{}, 40, theme))
	if !strings.Contains(out, "Error:") {
		t.Fatalf("expected error output")
	}
	out = stripANSI(renderReviews("Title", []letterboxd.Review{{Author: "Jane", Text: "Nice"}}, nil, 40, theme))
	if !strings.Contains(out, "Jane") || !strings.Contains(out, "Nice") {
		t.Fatalf("unexpected reviews output: %q", out)
	}
}

func TestRenderActivity(t *testing.T) {
	theme := newTheme()
	out := stripANSI(renderActivity(nil, errDummy{}, 0, 80, theme))
	if !strings.Contains(out, "Error:") {
		t.Fatalf("expected error output")
	}
	item := letterboxd.ActivityItem{
		When:  "2024-01-05T00:00:00Z",
		Title: "Inception",
		Parts: []letterboxd.SummaryPart{{Text: "Jane", Kind: "user"}, {Text: "watched", Kind: "text"}, {Text: "Inception", Kind: "movie"}},
	}
	out = stripANSI(renderActivity([]letterboxd.ActivityItem{item}, nil, 0, 80, theme))
	if !strings.Contains(out, "Inception") {
		t.Fatalf("unexpected activity output: %q", out)
	}
}

func TestRenderSummary(t *testing.T) {
	theme := newTheme()
	item := letterboxd.ActivityItem{
		Summary: "Jane watched Inception",
		Actor:   "Jane",
		Title:   "Inception",
		Rating:  "★★★★",
		Kind:    "diary",
	}
	out := stripANSI(renderSummary(item, theme))
	if !strings.Contains(out, "Jane") || !strings.Contains(out, "Inception") {
		t.Fatalf("unexpected summary output: %q", out)
	}
	item.Parts = []letterboxd.SummaryPart{{Text: "Jane", Kind: "user"}, {Text: "Inception", Kind: "movie"}}
	out = stripANSI(renderSummary(item, theme))
	if !strings.Contains(out, "Jane") || !strings.Contains(out, "Inception") {
		t.Fatalf("unexpected summary parts output: %q", out)
	}
}

func TestDescribeKind(t *testing.T) {
	if describeKind("watchlist") != "added to watchlist" {
		t.Fatalf("unexpected kind description")
	}
	if describeKind("unknown") != "" {
		t.Fatalf("expected empty description")
	}
}

func TestRenderBreadcrumbs(t *testing.T) {
	out := stripANSI(renderBreadcrumbs([]string{"alice"}, "bob", newTheme().user))
	if !strings.Contains(out, "@alice") || !strings.Contains(out, "@bob") {
		t.Fatalf("unexpected breadcrumbs: %q", out)
	}
}

type errDummy struct{}

func (errDummy) Error() string { return "boom" }
