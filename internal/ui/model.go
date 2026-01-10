package ui

import (
	"github.com/charmbracelet/bubbles/viewport"

	"letterboxd-tui/internal/letterboxd"
)

type tab int

const (
	tabProfile tab = iota
	tabDiary
	tabWatchlist
	tabFilm
	tabFollowing
	tabActivity
)

type listState struct {
	selected int
}

type Model struct {
	username     string
	profileUser  string
	profileStack []string
	client       *letterboxd.Client
	width        int
	height       int
	activeTab    tab
	lastTab      tab
	profile      letterboxd.Profile
	diary        []letterboxd.DiaryEntry
	watchlist    []letterboxd.WatchlistItem
	activity     []letterboxd.ActivityItem
	following    []letterboxd.ActivityItem
	film         letterboxd.Film
	profileErr   error
	diaryErr     error
	watchErr     error
	activityErr  error
	followErr    error
	filmErr      error
	loading      bool
	diaryList    listState
	watchList    listState
	actList      listState
	followList   listState
	viewport     viewport.Model
	modalVP      viewport.Model
	filmReturn   tab
	profileModal bool
}

func NewModel(username string, client *letterboxd.Client) Model {
	return Model{
		username:    username,
		profileUser: username,
		client:      client,
		activeTab:   tabProfile,
		lastTab:     tabProfile,
		loading:     true,
		viewport:    viewport.New(0, 0),
		modalVP:     viewport.New(0, 0),
	}
}

func (m *Model) moveSelection(delta int) {
	switch m.activeTab {
	case tabDiary:
		if len(m.diary) == 0 {
			return
		}
		m.diaryList.selected = clamp(m.diaryList.selected+delta, 0, len(m.diary)-1)
	case tabWatchlist:
		if len(m.watchlist) == 0 {
			return
		}
		m.watchList.selected = clamp(m.watchList.selected+delta, 0, len(m.watchlist)-1)
	case tabActivity:
		if len(m.activity) == 0 {
			return
		}
		m.actList.selected = clamp(m.actList.selected+delta, 0, len(m.activity)-1)
	case tabFollowing:
		if len(m.following) == 0 {
			return
		}
		m.followList.selected = clamp(m.followList.selected+delta, 0, len(m.following)-1)
	}
}

func (m *Model) resetTabPosition() {
	if m.activeTab == m.lastTab {
		return
	}
	m.viewport.YOffset = 0
	switch m.activeTab {
	case tabDiary:
		m.diaryList.selected = 0
	case tabWatchlist:
		m.watchList.selected = 0
	case tabFilm:
		m.filmErr = nil
	case tabActivity:
		m.actList.selected = 0
	case tabFollowing:
		m.followList.selected = 0
	}
	m.lastTab = m.activeTab
}

func (m *Model) pageSelection(dir int) {
	step := max(1, m.viewport.Height-1)
	switch m.activeTab {
	case tabDiary:
		if len(m.diary) == 0 {
			return
		}
		m.diaryList.selected = clamp(m.diaryList.selected+dir*step, 0, len(m.diary)-1)
	case tabWatchlist:
		if len(m.watchlist) == 0 {
			return
		}
		m.watchList.selected = clamp(m.watchList.selected+dir*step, 0, len(m.watchlist)-1)
	case tabActivity:
		if len(m.activity) == 0 {
			return
		}
		m.actList.selected = clamp(m.actList.selected+dir*step, 0, len(m.activity)-1)
	case tabFollowing:
		if len(m.following) == 0 {
			return
		}
		m.followList.selected = clamp(m.followList.selected+dir*step, 0, len(m.following)-1)
	}
}

func (m *Model) syncViewportToSelection() {
	var total, selected int
	switch m.activeTab {
	case tabDiary:
		total = len(m.diary)
		selected = m.diaryList.selected
	case tabWatchlist:
		total = len(m.watchlist)
		selected = m.watchList.selected
	case tabActivity:
		total = len(m.activity)
		selected = m.actList.selected
	case tabFollowing:
		total = len(m.following)
		selected = m.followList.selected
	default:
		return
	}
	if total == 0 || m.viewport.Height <= 0 {
		return
	}
	top := m.viewport.YOffset
	bottom := top + m.viewport.Height - 1
	if selected < top {
		m.viewport.YOffset = selected
	} else if selected > bottom {
		m.viewport.YOffset = selected - m.viewport.Height + 1
	}
	if m.viewport.YOffset < 0 {
		m.viewport.YOffset = 0
	}
	maxOffset := max(0, total-m.viewport.Height)
	if m.viewport.YOffset > maxOffset {
		m.viewport.YOffset = maxOffset
	}
}

func (m Model) openSelectedProfile() Model {
	if m.activeTab != tabFollowing || len(m.following) == 0 {
		return m
	}
	item := m.following[m.followList.selected]
	username := letterboxd.UsernameFromURL(item.ActorURL)
	if username == "" {
		username = letterboxd.UsernameFromURL(item.FilmURL)
	}
	if username == "" {
		return m
	}
	m.profileUser = username
	m.loading = true
	m.profileModal = true
	m.modalVP.YOffset = 0
	return m
}

func (m Model) openSelectedFilm() Model {
	var filmURL string
	switch m.activeTab {
	case tabDiary:
		if len(m.diary) == 0 {
			return m
		}
		filmURL = m.diary[m.diaryList.selected].FilmURL
	case tabWatchlist:
		if len(m.watchlist) == 0 {
			return m
		}
		filmURL = m.watchlist[m.watchList.selected].FilmURL
	case tabActivity:
		if len(m.activity) == 0 {
			return m
		}
		filmURL = m.activity[m.actList.selected].FilmURL
	case tabFollowing:
		if len(m.following) == 0 {
			return m
		}
		filmURL = m.following[m.followList.selected].FilmURL
	}
	filmURL = letterboxd.NormalizeFilmURL(filmURL)
	if filmURL == "" {
		return m
	}
	m.film = letterboxd.Film{URL: filmURL}
	m.filmReturn = m.activeTab
	m.activeTab = tabFilm
	m.loading = true
	m.viewport.YOffset = 0
	return m
}

func (m Model) goBackProfile() Model {
	if len(m.profileStack) == 0 {
		return m
	}
	last := m.profileStack[len(m.profileStack)-1]
	m.profileStack = m.profileStack[:len(m.profileStack)-1]
	m.profileUser = last
	m.activeTab = tabProfile
	m.loading = true
	return m
}

func (m Model) modalOpen() bool {
	return m.activeTab == tabFilm || m.profileModal
}
