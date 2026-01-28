package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/solean/letterboxd-tui/internal/letterboxd"
)

func TestUpdateWindowSize(t *testing.T) {
	m := NewModel("jane", nil)
	model, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	out := model.(Model)
	if out.width != 100 || out.height != 40 {
		t.Fatalf("unexpected size: %dx%d", out.width, out.height)
	}
	if out.searchInput.Width == 0 {
		t.Fatalf("expected search input width")
	}
}

func TestHandleSearchKeyEnter(t *testing.T) {
	m := NewModel("jane", nil)
	m.activeTab = tabSearch
	m.searchInput.SetValue("inception")
	cmd, handled := m.handleSearchKey(tea.KeyMsg{Type: tea.KeyEnter})
	if !handled || cmd == nil {
		t.Fatalf("expected handled enter with cmd")
	}
	if !m.searchLoading {
		t.Fatalf("expected searchLoading true")
	}
}

func TestHandleSearchKeyNavigation(t *testing.T) {
	m := NewModel("jane", nil)
	m.activeTab = tabSearch
	m.searchFocusInput = false
	m.searchResults = []letterboxd.SearchResult{{Title: "A"}, {Title: "B"}}
	_, handled := m.handleSearchKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if !handled || m.searchList.selected != 1 {
		t.Fatalf("expected selection to move")
	}
}

func TestUpdateLogModalToggle(t *testing.T) {
	m := NewModel("jane", nil)
	m.logModal = true
	m.logForm = newLogForm(letterboxd.Film{})
	m.logForm.focus = logFieldRewatch
	model, _ := m.updateLogModal(tea.KeyMsg{Type: tea.KeyEnter})
	out := model.(Model)
	if !out.logForm.rewatch {
		t.Fatalf("expected rewatch toggle")
	}
}

func TestUpdateWatchlistResult(t *testing.T) {
	m := NewModel("jane", nil)
	m.film.URL = letterboxd.BaseURL + "/film/inception/"
	model, _ := m.Update(watchlistResultMsg{inWatchlist: true})
	out := model.(Model)
	if !out.film.InWatchlist || out.watchlistStatus == "" {
		t.Fatalf("expected watchlist status update")
	}
}

func TestUpdateSearchMsg(t *testing.T) {
	m := NewModel("jane", nil)
	results := []letterboxd.SearchResult{{Title: "A"}}
	model, _ := m.Update(searchMsg{results: results})
	out := model.(Model)
	if len(out.searchResults) != 1 || out.searchLoading {
		t.Fatalf("unexpected search state")
	}
}

func TestUpdateActivityMsg(t *testing.T) {
	m := NewModel("jane", nil)
	items := []letterboxd.ActivityItem{{Title: "A"}}
	model, _ := m.Update(activityMsg{tab: tabActivity, items: items, after: ""})
	out := model.(Model)
	if len(out.activity) != 1 {
		t.Fatalf("unexpected activity state")
	}
}
