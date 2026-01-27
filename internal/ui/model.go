package ui

import (
	"errors"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"letterboxd-tui/internal/letterboxd"
)

type tab int

const (
	tabProfile tab = iota
	tabDiary
	tabWatchlist
	tabSearch
	tabFilm
	tabFollowing
	tabActivity
)

type listState struct {
	selected int
}

type Model struct {
	username                 string
	profileUser              string
	profileStack             []string
	client                   *letterboxd.Client
	width                    int
	height                   int
	activeTab                tab
	lastTab                  tab
	profile                  letterboxd.Profile
	diary                    []letterboxd.DiaryEntry
	watchlist                []letterboxd.WatchlistItem
	watchlistLoaded          bool
	activity                 []letterboxd.ActivityItem
	following                []letterboxd.ActivityItem
	searchResults            []letterboxd.SearchResult
	film                     letterboxd.Film
	modalProfile             letterboxd.Profile
	popReviews               []letterboxd.Review
	friendReviews            []letterboxd.Review
	popReviewsPage           int
	friendReviewsPage        int
	popReviewsDone           bool
	friendReviewsDone        bool
	popReviewsLoadingMore    bool
	friendReviewsLoadingMore bool
	popReviewsMoreErr        error
	friendReviewsMoreErr     error
	profileErr               error
	diaryErr                 error
	watchErr                 error
	activityErr              error
	followErr                error
	filmErr                  error
	popReviewsErr            error
	friendReviewsErr         error
	searchErr                error
	modalProfileErr          error
	loading                  bool
	modalLoading             bool
	searchLoading            bool
	diaryList                listState
	watchList                listState
	actList                  listState
	followList               listState
	searchList               listState
	diaryPage                int
	watchPage                int
	diaryLoadingMore         bool
	watchLoadingMore         bool
	activityLoadingMore      bool
	followLoadingMore        bool
	diaryDone                bool
	watchDone                bool
	activityDone             bool
	followDone               bool
	diaryMoreErr             error
	watchMoreErr             error
	activityMoreErr          error
	followMoreErr            error
	viewport                 viewport.Model
	modalVP                  viewport.Model
	filmReturn               tab
	profileModal             bool
	modalUser                string
	logModal                 bool
	logForm                  logForm
	logSpinner               spinner.Model
	watchlistStatus          string
	watchlistPending         bool
	searchInput              textinput.Model
	searchFocusInput         bool
	keys                     keyMap
	help                     help.Model
	pendingG                 bool
	pendingGAt               time.Time
}

func NewModel(username string, client *letterboxd.Client) Model {
	searchInput := textinput.New()
	searchInput.Placeholder = "Search films"
	searchInput.CharLimit = 80
	return Model{
		username:         username,
		profileUser:      username,
		client:           client,
		activeTab:        tabProfile,
		lastTab:          tabProfile,
		loading:          true,
		viewport:         viewport.New(0, 0),
		modalVP:          viewport.New(0, 0),
		logSpinner:       spinner.New(spinner.WithSpinner(spinner.Dot)),
		searchInput:      searchInput,
		searchFocusInput: true,
		keys:             newKeyMap(),
		help:             help.New(),
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
	case tabSearch:
		if len(m.searchResults) == 0 {
			return
		}
		m.searchList.selected = clamp(m.searchList.selected+delta, 0, len(m.searchResults)-1)
	}
}

func (m *Model) resetTabPosition() {
	if m.activeTab == m.lastTab {
		return
	}
	m.viewport.YOffset = 0
	m.searchInput.Blur()
	m.searchFocusInput = false
	switch m.activeTab {
	case tabDiary:
		m.diaryList.selected = 0
	case tabWatchlist:
		m.watchList.selected = 0
	case tabSearch:
		m.searchList.selected = 0
		m.searchFocusInput = true
		m.searchInput.Focus()
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
	case tabSearch:
		if len(m.searchResults) == 0 {
			return
		}
		m.searchList.selected = clamp(m.searchList.selected+dir*step, 0, len(m.searchResults)-1)
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
	case tabSearch:
		total = len(m.searchResults)
		selected = m.searchList.selected
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

func (m *Model) jumpToTop() {
	if m.profileModal || m.activeTab == tabFilm {
		m.modalVP.GotoTop()
		return
	}
	switch m.activeTab {
	case tabProfile:
		m.viewport.GotoTop()
	case tabDiary:
		if len(m.diary) == 0 {
			return
		}
		m.diaryList.selected = 0
		m.syncViewportToSelection()
	case tabWatchlist:
		if len(m.watchlist) == 0 {
			return
		}
		m.watchList.selected = 0
		m.syncViewportToSelection()
	case tabActivity:
		if len(m.activity) == 0 {
			return
		}
		m.actList.selected = 0
		m.syncViewportToSelection()
	case tabFollowing:
		if len(m.following) == 0 {
			return
		}
		m.followList.selected = 0
		m.syncViewportToSelection()
	case tabSearch:
		if len(m.searchResults) == 0 {
			return
		}
		m.searchList.selected = 0
		m.syncViewportToSelection()
	default:
		m.viewport.GotoTop()
	}
}

func (m *Model) jumpToBottom() {
	if m.profileModal || m.activeTab == tabFilm {
		m.modalVP.GotoBottom()
		return
	}
	switch m.activeTab {
	case tabProfile:
		m.viewport.GotoBottom()
	case tabDiary:
		if len(m.diary) == 0 {
			return
		}
		m.diaryList.selected = len(m.diary) - 1
		m.syncViewportToSelection()
	case tabWatchlist:
		if len(m.watchlist) == 0 {
			return
		}
		m.watchList.selected = len(m.watchlist) - 1
		m.syncViewportToSelection()
	case tabActivity:
		if len(m.activity) == 0 {
			return
		}
		m.actList.selected = len(m.activity) - 1
		m.syncViewportToSelection()
	case tabFollowing:
		if len(m.following) == 0 {
			return
		}
		m.followList.selected = len(m.following) - 1
		m.syncViewportToSelection()
	case tabSearch:
		if len(m.searchResults) == 0 {
			return
		}
		m.searchList.selected = len(m.searchResults) - 1
		m.syncViewportToSelection()
	default:
		m.viewport.GotoBottom()
	}
}

func (m *Model) resetPagination() {
	m.diaryPage = 0
	m.watchPage = 0
	m.diaryLoadingMore = false
	m.watchLoadingMore = false
	m.activityLoadingMore = false
	m.followLoadingMore = false
	m.diaryDone = false
	m.watchDone = false
	m.activityDone = false
	m.followDone = false
	m.diaryMoreErr = nil
	m.watchMoreErr = nil
	m.activityMoreErr = nil
	m.followMoreErr = nil
}

func (m *Model) maybeLoadMoreCmd() tea.Cmd {
	const threshold = 3
	switch m.activeTab {
	case tabDiary:
		if m.diaryLoadingMore || m.diaryDone || m.diaryMoreErr != nil {
			return nil
		}
		if len(m.diary) == 0 || m.diaryPage == 0 {
			return nil
		}
		if m.diaryList.selected < len(m.diary)-1-threshold {
			return nil
		}
		page := m.diaryPage + 1
		if page < 2 {
			page = 2
		}
		m.diaryLoadingMore = true
		return fetchDiaryCmd(m.client, m.username, page)
	case tabWatchlist:
		if m.watchLoadingMore || m.watchDone || m.watchMoreErr != nil {
			return nil
		}
		if len(m.watchlist) == 0 || m.watchPage == 0 {
			return nil
		}
		if m.watchList.selected < len(m.watchlist)-1-threshold {
			return nil
		}
		page := m.watchPage + 1
		if page < 2 {
			page = 2
		}
		m.watchLoadingMore = true
		return fetchWatchlistCmd(m.client, m.username, page)
	case tabActivity:
		if m.activityLoadingMore || m.activityDone || m.activityMoreErr != nil {
			return nil
		}
		if len(m.activity) == 0 {
			return nil
		}
		if m.actList.selected < len(m.activity)-1-threshold {
			return nil
		}
		after := lastActivityID(m.activity)
		if after == "" {
			return nil
		}
		m.activityLoadingMore = true
		return fetchActivityCmd(m.client, m.username, tabActivity, after)
	case tabFollowing:
		if m.followLoadingMore || m.followDone || m.followMoreErr != nil {
			return nil
		}
		if len(m.following) == 0 {
			return nil
		}
		if m.followList.selected < len(m.following)-1-threshold {
			return nil
		}
		after := lastActivityID(m.following)
		if after == "" {
			return nil
		}
		m.followLoadingMore = true
		return fetchActivityCmd(m.client, m.username, tabFollowing, after)
	}
	return nil
}

func (m *Model) maybeFillCmd() tea.Cmd {
	if m.modalOpen() || m.viewport.Height <= 0 {
		return nil
	}
	switch m.activeTab {
	case tabDiary:
		if m.diaryLoadingMore || m.diaryDone || m.diaryMoreErr != nil {
			return nil
		}
		if m.diaryPage == 0 || len(m.diary) == 0 {
			return nil
		}
		if len(m.diary) >= m.viewport.Height {
			return nil
		}
		page := m.diaryPage + 1
		if page < 2 {
			page = 2
		}
		m.diaryLoadingMore = true
		return fetchDiaryCmd(m.client, m.username, page)
	case tabWatchlist:
		if m.watchLoadingMore || m.watchDone || m.watchMoreErr != nil {
			return nil
		}
		if m.watchPage == 0 || len(m.watchlist) == 0 {
			return nil
		}
		if len(m.watchlist) >= m.viewport.Height {
			return nil
		}
		page := m.watchPage + 1
		if page < 2 {
			page = 2
		}
		m.watchLoadingMore = true
		return fetchWatchlistCmd(m.client, m.username, page)
	case tabActivity:
		if m.activityLoadingMore || m.activityDone || m.activityMoreErr != nil {
			return nil
		}
		if len(m.activity) == 0 {
			return nil
		}
		if len(m.activity) >= m.viewport.Height {
			return nil
		}
		after := lastActivityID(m.activity)
		if after == "" {
			return nil
		}
		m.activityLoadingMore = true
		return fetchActivityCmd(m.client, m.username, tabActivity, after)
	case tabFollowing:
		if m.followLoadingMore || m.followDone || m.followMoreErr != nil {
			return nil
		}
		if len(m.following) == 0 {
			return nil
		}
		if len(m.following) >= m.viewport.Height {
			return nil
		}
		after := lastActivityID(m.following)
		if after == "" {
			return nil
		}
		m.followLoadingMore = true
		return fetchActivityCmd(m.client, m.username, tabFollowing, after)
	}
	return nil
}

func (m Model) hasPopularReviewsSection() bool {
	return len(m.popReviews) > 0 || m.popReviewsErr != nil || m.popReviewsLoadingMore || m.popReviewsMoreErr != nil
}

func (m Model) hasFriendReviewsSection() bool {
	return len(m.friendReviews) > 0 || m.friendReviewsErr != nil || m.friendReviewsLoadingMore || m.friendReviewsMoreErr != nil
}

func (m *Model) maybeLoadMoreReviewsCmd() tea.Cmd {
	if m.activeTab != tabFilm || m.logModal {
		return nil
	}
	if m.film.Slug == "" {
		return nil
	}
	if m.modalVP.Height <= 0 {
		return nil
	}
	if m.modalVP.ScrollPercent() < 0.85 {
		return nil
	}

	kind := ""
	if m.hasPopularReviewsSection() {
		kind = "popular"
	} else if m.hasFriendReviewsSection() {
		kind = "friends"
	}
	if kind == "" {
		return nil
	}

	switch kind {
	case "popular":
		if m.popReviewsLoadingMore || m.popReviewsDone || m.popReviewsMoreErr != nil {
			return nil
		}
		if m.popReviewsPage == 0 || len(m.popReviews) == 0 {
			return nil
		}
		page := m.popReviewsPage + 1
		if page < 2 {
			page = 2
		}
		m.popReviewsLoadingMore = true
		m.refreshModalViewport()
		return fetchReviewsCmd(m.client, m.film.Slug, "popular", page)
	case "friends":
		if m.friendReviewsLoadingMore || m.friendReviewsDone || m.friendReviewsMoreErr != nil {
			return nil
		}
		if m.friendReviewsPage == 0 || len(m.friendReviews) == 0 {
			return nil
		}
		page := m.friendReviewsPage + 1
		if page < 2 {
			page = 2
		}
		m.friendReviewsLoadingMore = true
		m.refreshModalViewport()
		return fetchReviewsCmd(m.client, m.film.Slug, "friends", page)
	}
	return nil
}

func lastActivityID(items []letterboxd.ActivityItem) string {
	for i := len(items) - 1; i >= 0; i-- {
		if items[i].ID != "" {
			return items[i].ID
		}
	}
	return ""
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
	m.modalUser = username
	m.modalProfile = letterboxd.Profile{}
	m.modalProfileErr = nil
	m.modalLoading = true
	m.profileModal = true
	m.modalVP.YOffset = 0
	m.refreshModalViewport()
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
	case tabSearch:
		if len(m.searchResults) == 0 {
			return m
		}
		filmURL = m.searchResults[m.searchList.selected].FilmURL
	}
	filmURL = letterboxd.NormalizeFilmURL(filmURL)
	if filmURL == "" {
		return m
	}
	m.film = letterboxd.Film{URL: filmURL}
	m.filmErr = nil
	m.popReviews = nil
	m.friendReviews = nil
	m.popReviewsPage = 0
	m.friendReviewsPage = 0
	m.popReviewsDone = false
	m.friendReviewsDone = false
	m.popReviewsLoadingMore = false
	m.friendReviewsLoadingMore = false
	m.popReviewsMoreErr = nil
	m.friendReviewsMoreErr = nil
	m.popReviewsErr = nil
	m.friendReviewsErr = nil
	m.watchlistPending = false
	m.watchlistStatus = ""
	m.filmReturn = m.activeTab
	m.activeTab = tabFilm
	m.loading = true
	m.viewport.YOffset = 0
	m.modalVP.YOffset = 0
	m.modalVP.SetContent("")
	m.refreshModalViewport()
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

func (m Model) startLogModal() Model {
	if m.film.ViewingUID == "" {
		m.filmErr = errors.New("cannot log this film (missing id)")
		return m
	}
	form := newLogForm(m.film)
	form.setSize(m.width)
	m.logForm = form
	m.logModal = true
	return m
}

func (m Model) buildDiaryRequest() letterboxd.DiaryEntryRequest {
	ratingStr := strings.TrimSpace(m.logForm.rating.Value())
	ratingVal := 0
	if ratingStr != "" {
		if val, err := strconv.ParseFloat(ratingStr, 64); err == nil {
			ratingVal = int(math.Round(val * 2))
		}
	}
	req := letterboxd.DiaryEntryRequest{
		ViewingUID:       m.film.ViewingUID,
		WatchedDate:      strings.TrimSpace(m.logForm.date.Value()),
		RatingValue:      clamp(ratingVal, 0, 10),
		Review:           m.logForm.review.Value(),
		ContainsSpoilers: m.logForm.spoilers,
		Rewatch:          m.logForm.rewatch,
		Tags:             strings.TrimSpace(m.logForm.tags.Value()),
		Liked:            m.logForm.liked,
		Privacy:          "",
		Draft:            m.logForm.draft,
		Referer:          m.film.URL,
		JSONResponse:     true,
	}
	if req.WatchedDate == "" {
		req.WatchedDate = time.Now().Format("2006-01-02")
	}
	priv := m.logForm.privacyLabel()
	if priv != "" && priv != "Default" {
		req.Privacy = priv
	}
	return req
}

func (m Model) buildWatchlistRequest() (letterboxd.WatchlistRequest, error) {
	slug := strings.TrimSpace(m.film.Slug)
	if slug == "" {
		slug = letterboxd.FilmSlug(m.film.URL)
	}
	watchlistID := strings.TrimSpace(m.film.WatchlistID)
	if watchlistID == "" && slug == "" && strings.TrimSpace(m.film.FilmID) == "" {
		return letterboxd.WatchlistRequest{}, errors.New("missing film id")
	}
	req := letterboxd.WatchlistRequest{
		WatchlistID:  watchlistID,
		FilmID:       m.film.FilmID,
		FilmSlug:     slug,
		Referer:      m.film.URL,
		JSONResponse: true,
	}
	return req, nil
}

func (m Model) watchlistState() (bool, bool) {
	if m.film.WatchlistOK {
		return m.film.InWatchlist, true
	}
	if !m.watchlistLoaded || m.watchErr != nil {
		return false, false
	}
	filmURL := letterboxd.NormalizeFilmURL(m.film.URL)
	if filmURL == "" {
		return false, false
	}
	for _, item := range m.watchlist {
		if letterboxd.NormalizeFilmURL(item.FilmURL) == filmURL {
			return true, true
		}
	}
	return false, true
}

func (m Model) modalOpen() bool {
	return m.activeTab == tabFilm || m.profileModal || m.logModal
}
