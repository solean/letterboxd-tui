package ui

import (
	"testing"
	"time"

	"letterboxd-tui/internal/letterboxd"
)

func TestMoveSelectionAndPageSelection(t *testing.T) {
	m := NewModel("jane", nil)
	m.activeTab = tabDiary
	m.diary = []letterboxd.DiaryEntry{{Title: "A"}, {Title: "B"}, {Title: "C"}}
	m.moveSelection(1)
	if m.diaryList.selected != 1 {
		t.Fatalf("unexpected selection: %d", m.diaryList.selected)
	}
	m.pageSelection(1)
	if m.diaryList.selected <= 1 {
		t.Fatalf("expected selection to move, got %d", m.diaryList.selected)
	}
}

func TestResetTabPosition(t *testing.T) {
	m := NewModel("jane", nil)
	m.activeTab = tabSearch
	m.lastTab = tabProfile
	m.searchFocusInput = false
	m.resetTabPosition()
	if !m.searchFocusInput {
		t.Fatalf("expected search focus")
	}
	if m.lastTab != tabSearch {
		t.Fatalf("expected last tab update")
	}
}

func TestSyncViewportToSelection(t *testing.T) {
	m := NewModel("jane", nil)
	m.activeTab = tabWatchlist
	m.watchlist = make([]letterboxd.WatchlistItem, 10)
	m.watchList.selected = 8
	m.viewport.Height = 3
	m.syncViewportToSelection()
	if m.viewport.YOffset == 0 {
		t.Fatalf("expected viewport to move")
	}
}

func TestOpenSelectedProfile(t *testing.T) {
	m := NewModel("jane", nil)
	m.activeTab = tabFollowing
	m.following = []letterboxd.ActivityItem{{ActorURL: letterboxd.BaseURL + "/alice/"}}
	m.followList.selected = 0
	m = m.openSelectedProfile()
	if !m.profileModal || m.modalUser != "alice" {
		t.Fatalf("unexpected profile modal state: %+v", m)
	}
}

func TestOpenSelectedFilm(t *testing.T) {
	m := NewModel("jane", nil)
	m.activeTab = tabDiary
	m.diary = []letterboxd.DiaryEntry{{FilmURL: letterboxd.BaseURL + "/film/inception/"}}
	m = m.openSelectedFilm()
	if m.activeTab != tabFilm || m.film.URL == "" {
		t.Fatalf("expected film modal, got %+v", m)
	}
}

func TestGoBackProfile(t *testing.T) {
	m := NewModel("jane", nil)
	m.profileStack = []string{"alice"}
	m = m.goBackProfile()
	if m.profileUser != "alice" {
		t.Fatalf("unexpected profile user: %q", m.profileUser)
	}
}

func TestStartLogModal(t *testing.T) {
	m := NewModel("jane", nil)
	m = m.startLogModal()
	if m.filmErr == nil {
		t.Fatalf("expected error for missing viewing id")
	}
	m.film.ViewingUID = "film:123"
	m = m.startLogModal()
	if !m.logModal || !m.logForm.active {
		t.Fatalf("expected log modal")
	}
}

func TestBuildDiaryRequest(t *testing.T) {
	m := NewModel("jane", nil)
	m.film.ViewingUID = "film:123"
	m.film.URL = letterboxd.BaseURL + "/film/inception/"
	m.logForm = newLogForm(letterboxd.Film{})
	m.logForm.rating.SetValue("4.5")
	m.logForm.date.SetValue("")
	m.logForm.tags.SetValue("tag1")
	m.logForm.review.SetValue("Nice")
	m.logForm.privacyIndex = 1
	req := m.buildDiaryRequest()
	if req.RatingValue != 9 {
		t.Fatalf("unexpected rating value: %d", req.RatingValue)
	}
	if req.WatchedDate == "" {
		t.Fatalf("expected watched date")
	}
	if _, err := time.Parse("2006-01-02", req.WatchedDate); err != nil {
		t.Fatalf("expected date format, got %q", req.WatchedDate)
	}
	if req.Privacy != "Anyone" {
		t.Fatalf("unexpected privacy: %q", req.Privacy)
	}
}

func TestBuildWatchlistRequest(t *testing.T) {
	m := NewModel("jane", nil)
	m.film = letterboxd.Film{URL: letterboxd.BaseURL + "/film/inception/", FilmID: "123"}
	req, err := m.buildWatchlistRequest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.FilmSlug != "inception" || req.FilmID != "123" {
		t.Fatalf("unexpected request: %+v", req)
	}
	m.film = letterboxd.Film{}
	if _, err := m.buildWatchlistRequest(); err == nil {
		t.Fatalf("expected error for missing film id")
	}
}

func TestWatchlistState(t *testing.T) {
	m := NewModel("jane", nil)
	m.film = letterboxd.Film{URL: letterboxd.BaseURL + "/film/inception/"}
	m.watchlistLoaded = true
	m.watchlist = []letterboxd.WatchlistItem{{FilmURL: letterboxd.BaseURL + "/film/inception/"}}
	if in, ok := m.watchlistState(); !ok || !in {
		t.Fatalf("expected in watchlist")
	}
	m.film.WatchlistOK = true
	m.film.InWatchlist = false
	if in, ok := m.watchlistState(); !ok || in {
		t.Fatalf("unexpected watchlist state")
	}
}

func TestModalOpen(t *testing.T) {
	m := NewModel("jane", nil)
	m.activeTab = tabFilm
	if !m.modalOpen() {
		t.Fatalf("expected modal open")
	}
	m.activeTab = tabProfile
	if m.modalOpen() {
		t.Fatalf("expected modal closed")
	}
	m.profileModal = true
	if !m.modalOpen() {
		t.Fatalf("expected modal open")
	}
}
