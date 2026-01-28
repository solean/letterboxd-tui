package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/solean/letterboxd-tui/internal/letterboxd"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = ws.Width
		m.height = ws.Height
		bodyHeight := max(1, m.height-3)
		m.viewport.Width = m.width
		m.viewport.Height = bodyHeight
		m.help.Width = m.width
		if m.logModal {
			m.logForm.setSize(m.width)
		}
		m.searchInput.Width = max(10, m.width-4)
		m.refreshModalViewport()
		cmd := m.maybeFillCmd()
		if m.activeTab == tabFilm {
			cmd = tea.Batch(cmd, m.maybeLoadMoreReviewsCmd())
		}
		return m, cmd
	}

	switch sm := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.logSpinner, cmd = m.logSpinner.Update(sm)
		if m.logModal {
			return m, cmd
		}
		return m, nil
	}

	if rm, ok := msg.(reviewsMsg); ok {
		switch rm.kind {
		case "popular":
			if rm.page <= 1 {
				m.popReviews = rm.reviews
				m.popReviewsErr = rm.err
				m.popReviewsPage = max(1, rm.page)
				m.popReviewsDone = rm.err == nil && len(rm.reviews) == 0
				m.popReviewsLoadingMore = false
				m.popReviewsMoreErr = nil
			} else {
				m.popReviewsLoadingMore = false
				if rm.err != nil {
					m.popReviewsMoreErr = rm.err
					m.refreshModalViewport()
					return m, nil
				}
				m.popReviewsMoreErr = nil
				var added int
				m.popReviews, added = appendReviews(m.popReviews, rm.reviews)
				if added == 0 {
					m.popReviewsDone = true
				} else {
					m.popReviewsPage = rm.page
				}
			}
		case "friends":
			if rm.page <= 1 {
				m.friendReviews = rm.reviews
				m.friendReviewsErr = rm.err
				m.friendReviewsPage = max(1, rm.page)
				m.friendReviewsDone = rm.err == nil && len(rm.reviews) == 0
				m.friendReviewsLoadingMore = false
				m.friendReviewsMoreErr = nil
			} else {
				m.friendReviewsLoadingMore = false
				if rm.err != nil {
					m.friendReviewsMoreErr = rm.err
					m.refreshModalViewport()
					return m, nil
				}
				m.friendReviewsMoreErr = nil
				var added int
				m.friendReviews, added = appendReviews(m.friendReviews, rm.reviews)
				if added == 0 {
					m.friendReviewsDone = true
				} else {
					m.friendReviewsPage = rm.page
				}
			}
		default:
			return m, nil
		}
		m.refreshModalViewport()
		return m, m.maybeLoadMoreReviewsCmd()
	}

	if m.logModal {
		return m.updateLogModal(msg)
	}

	switch ev := msg.(type) {
	case tea.KeyMsg:
		if cmd, handled := m.handleSearchKey(ev); handled {
			return m, cmd
		}
		switch {
		case key.Matches(ev, m.keys.QuitAll):
			return m, tea.Quit
		case key.Matches(ev, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case m.handleJumpKeys(ev):
			return m, nil
		case m.modalOpen() && key.Matches(ev, m.keys.ModalBack):
			if m.activeTab == tabFilm {
				m.activeTab = m.filmReturn
				m.resetTabPosition()
			} else if m.profileModal {
				m.profileModal = false
			}
			return m, nil
		case key.Matches(ev, m.keys.SearchTab):
			if m.profileModal {
				m.profileModal = false
			}
			if m.activeTab != tabSearch {
				m.activeTab = tabSearch
				m.resetTabPosition()
				return m, nil
			}
			m.searchFocusInput = true
			m.searchInput.Focus()
			return m, nil
		case key.Matches(ev, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(ev, m.keys.NextTab):
			if m.activeTab == tabFilm {
				m.activeTab = m.filmReturn
			} else {
				m.activeTab = nextTab(m, m.activeTab)
			}
			m.resetTabPosition()
			return m, m.maybeFillCmd()
		case key.Matches(ev, m.keys.PrevTab):
			if m.activeTab == tabFilm {
				m.activeTab = m.filmReturn
			} else {
				m.activeTab = prevTab(m, m.activeTab)
			}
			m.resetTabPosition()
			return m, m.maybeFillCmd()
		case key.Matches(ev, m.keys.Down):
			if m.modalOpen() {
				m.modalVP.LineDown(1)
				if m.activeTab == tabFilm {
					return m, m.maybeLoadMoreReviewsCmd()
				}
			} else if m.activeTab == tabProfile {
				m.viewport.LineDown(1)
			} else {
				m.moveSelection(1)
				m.syncViewportToSelection()
				return m, m.maybeLoadMoreCmd()
			}
		case key.Matches(ev, m.keys.Up):
			if m.modalOpen() {
				m.modalVP.LineUp(1)
			} else if m.activeTab == tabProfile {
				m.viewport.LineUp(1)
			} else {
				m.moveSelection(-1)
				m.syncViewportToSelection()
				return m, m.maybeLoadMoreCmd()
			}
		case key.Matches(ev, m.keys.PageDown):
			if m.modalOpen() {
				m.modalVP.ViewDown()
				if m.activeTab == tabFilm {
					return m, m.maybeLoadMoreReviewsCmd()
				}
			} else if m.activeTab == tabProfile {
				m.viewport.ViewDown()
			} else {
				m.pageSelection(1)
				m.syncViewportToSelection()
				return m, m.maybeLoadMoreCmd()
			}
		case key.Matches(ev, m.keys.PageUp):
			if m.modalOpen() {
				m.modalVP.ViewUp()
			} else if m.activeTab == tabProfile {
				m.viewport.ViewUp()
			} else {
				m.pageSelection(-1)
				m.syncViewportToSelection()
				return m, m.maybeLoadMoreCmd()
			}
		case key.Matches(ev, m.keys.Refresh):
			m.loading = true
			m.resetPagination()
			cmds := []tea.Cmd{
				fetchProfileCmd(m.client, m.profileUser),
				fetchDiaryCmd(m.client, m.username, 1),
				fetchWatchlistCmd(m.client, m.username, 1),
				fetchActivityCmd(m.client, m.username, tabActivity, ""),
			}
			if m.hasCookie() {
				cmds = append(cmds, fetchActivityCmd(m.client, m.username, tabFollowing, ""))
			}
			return m, tea.Batch(cmds...)
		case key.Matches(ev, m.keys.Select):
			if m.activeTab == tabFollowing {
				m = m.openSelectedProfile()
				return m, fetchProfileModalCmd(m.client, m.modalUser)
			} else if m.activeTab == tabDiary || m.activeTab == tabWatchlist || m.activeTab == tabActivity {
				m = m.openSelectedFilm()
				if m.activeTab == tabFilm {
					return m, fetchFilmCmd(m.client, m.film.URL, m.username)
				}
			}
		case key.Matches(ev, m.keys.Back):
			if m.profileModal {
				m.profileModal = false
			} else if m.activeTab == tabProfile {
				m = m.goBackProfile()
				if m.activeTab == tabProfile {
					return m, fetchProfileCmd(m.client, m.profileUser)
				}
			}
		case key.Matches(ev, m.keys.Log):
			if m.activeTab == tabFilm && m.hasCookie() {
				m = m.startLogModal()
				return m, m.logSpinner.Tick
			}
		case key.Matches(ev, m.keys.WatchlistAdd):
			if m.activeTab == tabFilm && m.hasCookie() {
				if m.watchlistPending {
					return m, nil
				}
				if inWatchlist, ok := m.watchlistState(); ok && inWatchlist {
					return m, nil
				}
				req, err := m.buildWatchlistRequest()
				if err != nil {
					m.watchlistStatus = "Error: " + err.Error()
					return m, nil
				}
				m.watchlistPending = true
				m.watchlistStatus = "Adding to watchlist..."
				return m, setWatchlistCmd(m.client, req, true)
			}
		case key.Matches(ev, m.keys.WatchlistRemove):
			if m.activeTab == tabFilm && m.hasCookie() {
				if m.watchlistPending {
					return m, nil
				}
				inWatchlist, ok := m.watchlistState()
				if !ok || !inWatchlist {
					return m, nil
				}
				req, err := m.buildWatchlistRequest()
				if err != nil {
					m.watchlistStatus = "Error: " + err.Error()
					return m, nil
				}
				m.watchlistPending = true
				m.watchlistStatus = "Removing from watchlist..."
				return m, setWatchlistCmd(m.client, req, false)
			}
		case key.Matches(ev, m.keys.Open):
			if m.profileModal {
				return m, openBrowserCmd(letterboxd.ProfileURL(m.modalUser))
			} else if m.activeTab == tabProfile {
				return m, openBrowserCmd(letterboxd.ProfileURL(m.profileUser))
			} else if m.activeTab == tabFilm {
				return m, openBrowserCmd(m.film.URL)
			}
		case key.Matches(ev, m.keys.Cancel):
			if m.activeTab == tabFilm {
				m.activeTab = m.filmReturn
				m.resetTabPosition()
			} else if m.profileModal {
				m.profileModal = false
			}
		}
	case profileMsg:
		if ev.modal {
			m.modalProfile = ev.profile
			m.modalProfileErr = ev.err
			m.modalLoading = false
			m.refreshModalViewport()
		} else {
			m.profile = ev.profile
			m.profileErr = ev.err
			m.loading = false
		}
	case diaryMsg:
		if ev.page <= 1 {
			m.diary = ev.items
			m.diaryErr = ev.err
			m.diaryPage = max(1, ev.page)
			m.diaryDone = ev.err == nil && len(ev.items) == 0
			m.diaryLoadingMore = false
			m.diaryMoreErr = nil
			m.loading = false
		} else {
			m.diaryLoadingMore = false
			if ev.err != nil {
				m.diaryMoreErr = ev.err
				return m, nil
			}
			m.diaryMoreErr = nil
			var added int
			m.diary, added = appendDiaryEntries(m.diary, ev.items)
			if added == 0 {
				m.diaryDone = true
			} else {
				m.diaryPage = ev.page
			}
		}
		return m, m.maybeFillCmd()
	case watchlistMsg:
		if ev.page <= 1 {
			m.watchlist = ev.items
			m.watchErr = ev.err
			m.watchlistLoaded = true
			m.watchPage = max(1, ev.page)
			m.watchDone = ev.err == nil && len(ev.items) == 0
			m.watchLoadingMore = false
			m.watchMoreErr = nil
			m.loading = false
		} else {
			m.watchLoadingMore = false
			if ev.err != nil {
				m.watchMoreErr = ev.err
				return m, nil
			}
			m.watchMoreErr = nil
			var added int
			m.watchlist, added = appendWatchlistItems(m.watchlist, ev.items)
			if added == 0 {
				m.watchDone = true
			} else {
				m.watchPage = ev.page
			}
		}
		return m, m.maybeFillCmd()
	case filmMsg:
		m.film = ev.film
		m.filmErr = ev.err
		m.loading = false
		m.refreshModalViewport()
		if ev.film.Slug != "" {
			cmds := []tea.Cmd{fetchReviewsCmd(m.client, ev.film.Slug, "popular", 1)}
			if m.hasCookie() {
				cmds = append(cmds, fetchReviewsCmd(m.client, ev.film.Slug, "friends", 1))
			}
			return m, tea.Batch(cmds...)
		}
	case searchMsg:
		m.searchResults = ev.results
		m.searchErr = ev.err
		m.searchLoading = false
		m.searchList.selected = 0
		m.searchFocusInput = false
		m.syncViewportToSelection()
	case activityMsg:
		if ev.after == "" {
			if ev.tab == tabActivity {
				m.activity = ev.items
				m.activityErr = ev.err
				m.activityDone = ev.err == nil && len(ev.items) == 0
				m.activityLoadingMore = false
				m.activityMoreErr = nil
			} else {
				m.following = ev.items
				m.followErr = ev.err
				m.followDone = ev.err == nil && len(ev.items) == 0
				m.followLoadingMore = false
				m.followMoreErr = nil
			}
			m.loading = false
			return m, m.maybeFillCmd()
		}
		if ev.tab == tabActivity {
			m.activityLoadingMore = false
			if ev.err != nil {
				m.activityMoreErr = ev.err
				return m, nil
			}
			m.activityMoreErr = nil
			var added int
			m.activity, added = appendActivityItems(m.activity, ev.items)
			if added == 0 {
				m.activityDone = true
			}
		} else {
			m.followLoadingMore = false
			if ev.err != nil {
				m.followMoreErr = ev.err
				return m, nil
			}
			m.followMoreErr = nil
			var added int
			m.following, added = appendActivityItems(m.following, ev.items)
			if added == 0 {
				m.followDone = true
			}
		}
		return m, m.maybeFillCmd()
	case errMsg:
		m.diaryErr = ev.err
		m.loading = false
	case openMsg:
		if ev.err != nil {
			m.profileErr = ev.err
		}
	case watchlistResultMsg:
		m.watchlistPending = false
		if ev.err != nil {
			m.watchlistStatus = "Error: " + ev.err.Error()
			return m, nil
		}
		m.film.InWatchlist = ev.inWatchlist
		m.film.WatchlistOK = true
		if ev.inWatchlist {
			m.watchlistStatus = "Added to watchlist."
		} else {
			m.watchlistStatus = "Removed from watchlist."
		}
		m.refreshModalViewport()
		m.loading = true
		return m, tea.Batch(
			fetchFilmCmd(m.client, m.film.URL, m.username),
			fetchWatchlistCmd(m.client, m.username, 1),
		)
	}
	return m, nil
}

func (m *Model) refreshModalViewport() {
	if m.logModal {
		return
	}
	if !m.profileModal && m.activeTab != tabFilm {
		return
	}
	theme := newTheme()
	width, height := modalDimensions(m.width, m.height)
	innerWidth := width - 4
	innerHeight := height - 2
	legend := renderHelp(*m, theme, innerWidth)
	bodyHeight := max(1, innerHeight-lipgloss.Height(legend)-1)

	m.modalVP.Width = innerWidth
	m.modalVP.Height = bodyHeight

	if m.profileModal {
		content := renderProfileContent(m.modalProfile, m.modalProfileErr, m.modalLoading, m.modalUser, nil, theme)
		m.modalVP.SetContent(content)
		return
	}
	content := renderFilm(*m, theme)
	m.modalVP.SetContent(content)
}

func (m *Model) handleSearchKey(msg tea.KeyMsg) (tea.Cmd, bool) {
	if m.activeTab != tabSearch {
		return nil, false
	}
	switch {
	case key.Matches(msg, m.keys.QuitAll, m.keys.Quit, m.keys.Help, m.keys.NextTab, m.keys.PrevTab):
		return nil, false
	case key.Matches(msg, m.keys.Select):
		if m.searchFocusInput {
			query := strings.TrimSpace(m.searchInput.Value())
			if query == "" {
				return nil, true
			}
			m.searchLoading = true
			m.searchErr = nil
			m.searchResults = nil
			m.searchFocusInput = false
			m.searchInput.Blur()
			return fetchSearchCmd(m.client, query), true
		}
		updated := m.openSelectedFilm()
		*m = updated
		if m.activeTab == tabFilm {
			return fetchFilmCmd(m.client, m.film.URL, m.username), true
		}
		return nil, true
	case key.Matches(msg, m.keys.SearchTab):
		m.searchFocusInput = true
		m.searchInput.Focus()
		return nil, true
	case key.Matches(msg, m.keys.Cancel):
		m.searchFocusInput = false
		m.searchInput.Blur()
		return nil, true
	case key.Matches(msg, m.keys.Down):
		if m.searchFocusInput {
			break
		}
		m.moveSelection(1)
		m.syncViewportToSelection()
		return nil, true
	case key.Matches(msg, m.keys.Up):
		if m.searchFocusInput {
			break
		}
		m.moveSelection(-1)
		m.syncViewportToSelection()
		return nil, true
	case key.Matches(msg, m.keys.PageDown):
		if m.searchFocusInput {
			break
		}
		m.pageSelection(1)
		m.syncViewportToSelection()
		return nil, true
	case key.Matches(msg, m.keys.PageUp):
		if m.searchFocusInput {
			break
		}
		m.pageSelection(-1)
		m.syncViewportToSelection()
		return nil, true
	}

	if m.searchFocusInput {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return cmd, true
	}
	return nil, false
}

func (m *Model) handleJumpKeys(msg tea.KeyMsg) bool {
	if key.Matches(msg, m.keys.JumpBottom) {
		m.pendingG = false
		m.jumpToBottom()
		return true
	}

	if msg.String() != "g" {
		m.pendingG = false
		return false
	}

	now := time.Now()
	if m.pendingG && now.Sub(m.pendingGAt) <= 600*time.Millisecond {
		m.pendingG = false
		m.jumpToTop()
		return true
	}

	m.pendingG = true
	m.pendingGAt = now
	return true
}

func (m Model) updateLogModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(typed, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(typed, m.keys.Cancel, m.keys.ModalBack):
			m.logModal = false
			m.logForm.submitting = false
			return m, nil
		case key.Matches(typed, m.keys.NextTab):
			m.logForm.focusField(m.logForm.focus + 1)
			return m, nil
		case key.Matches(typed, m.keys.PrevTab):
			m.logForm.focusField(m.logForm.focus - 1)
			return m, nil
		case key.Matches(typed, m.keys.Select, m.keys.Submit):
			if m.logForm.focus == logFieldSubmit {
				if m.film.ViewingUID == "" {
					m.logForm.status = "Missing film id; cannot log."
					return m, nil
				}
				req := m.buildDiaryRequest()
				m.logForm.submitting = true
				return m, saveDiaryEntryCmd(m.client, req)
			}
			switch m.logForm.focus {
			case logFieldRewatch:
				m.logForm.rewatch = !m.logForm.rewatch
				return m, nil
			case logFieldSpoilers:
				m.logForm.spoilers = !m.logForm.spoilers
				return m, nil
			case logFieldLiked:
				m.logForm.liked = !m.logForm.liked
				return m, nil
			case logFieldPrivacy:
				m.logForm.privacyIndex = (m.logForm.privacyIndex + 1) % len(privacyOptions)
				return m, nil
			case logFieldDraft:
				m.logForm.draft = !m.logForm.draft
				return m, nil
			}
		case key.Matches(typed, m.keys.Toggle):
			switch m.logForm.focus {
			case logFieldRewatch:
				m.logForm.rewatch = !m.logForm.rewatch
				return m, nil
			case logFieldSpoilers:
				m.logForm.spoilers = !m.logForm.spoilers
				return m, nil
			case logFieldLiked:
				m.logForm.liked = !m.logForm.liked
				return m, nil
			case logFieldPrivacy:
				m.logForm.privacyIndex = (m.logForm.privacyIndex + 1) % len(privacyOptions)
				return m, nil
			case logFieldDraft:
				m.logForm.draft = !m.logForm.draft
				return m, nil
			}
		}
	case logResultMsg:
		m.logForm.submitting = false
		if typed.err != nil {
			m.logForm.status = "Error: " + typed.err.Error()
			return m, nil
		}
		m.logForm.status = "Saved!"
		m.loading = true
		return m, fetchFilmCmd(m.client, m.film.URL, m.username)
	default:
		if m.logForm.submitting {
			var cmd tea.Cmd
			m.logSpinner, cmd = m.logSpinner.Update(msg)
			return m, cmd
		}
	}

	switch m.logForm.focus {
	case logFieldRating:
		m.logForm.rating, _ = m.logForm.rating.Update(msg)
	case logFieldDate:
		m.logForm.date, _ = m.logForm.date.Update(msg)
	case logFieldTags:
		m.logForm.tags, _ = m.logForm.tags.Update(msg)
	case logFieldReview:
		m.logForm.review, _ = m.logForm.review.Update(msg)
	}
	return m, nil
}

func appendDiaryEntries(existing, incoming []letterboxd.DiaryEntry) ([]letterboxd.DiaryEntry, int) {
	seen := make(map[string]struct{}, len(existing))
	for _, entry := range existing {
		key := diaryKey(entry)
		if key != "" {
			seen[key] = struct{}{}
		}
	}
	added := 0
	for _, entry := range incoming {
		key := diaryKey(entry)
		if key != "" {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
		}
		existing = append(existing, entry)
		added++
	}
	return existing, added
}

func diaryKey(entry letterboxd.DiaryEntry) string {
	if entry.FilmURL != "" {
		return entry.FilmURL + "|" + entry.Date
	}
	if entry.Title != "" {
		return entry.Title + "|" + entry.Date
	}
	return ""
}

func appendWatchlistItems(existing, incoming []letterboxd.WatchlistItem) ([]letterboxd.WatchlistItem, int) {
	seen := make(map[string]struct{}, len(existing))
	for _, item := range existing {
		key := watchlistKey(item)
		if key != "" {
			seen[key] = struct{}{}
		}
	}
	added := 0
	for _, item := range incoming {
		key := watchlistKey(item)
		if key != "" {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
		}
		existing = append(existing, item)
		added++
	}
	return existing, added
}

func watchlistKey(item letterboxd.WatchlistItem) string {
	if item.FilmURL != "" {
		return item.FilmURL
	}
	if item.Title != "" {
		return item.Title + "|" + item.Year
	}
	return ""
}

func appendActivityItems(existing, incoming []letterboxd.ActivityItem) ([]letterboxd.ActivityItem, int) {
	seen := make(map[string]struct{}, len(existing))
	for _, item := range existing {
		key := activityKey(item)
		if key != "" {
			seen[key] = struct{}{}
		}
	}
	added := 0
	for _, item := range incoming {
		key := activityKey(item)
		if key != "" {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
		}
		existing = append(existing, item)
		added++
	}
	return existing, added
}

func activityKey(item letterboxd.ActivityItem) string {
	if item.ID != "" {
		return item.ID
	}
	if item.FilmURL != "" || item.ActorURL != "" {
		return item.ActorURL + "|" + item.FilmURL + "|" + item.When
	}
	return item.Summary + "|" + item.When
}

func appendReviews(existing, incoming []letterboxd.Review) ([]letterboxd.Review, int) {
	seen := make(map[string]struct{}, len(existing))
	for _, review := range existing {
		key := reviewKey(review)
		if key != "" {
			seen[key] = struct{}{}
		}
	}
	added := 0
	for _, review := range incoming {
		key := reviewKey(review)
		if key != "" {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
		}
		existing = append(existing, review)
		added++
	}
	return existing, added
}

func reviewKey(review letterboxd.Review) string {
	if review.Link != "" {
		return review.Link
	}
	if review.Author != "" || review.Text != "" || review.Rating != "" {
		return review.Author + "|" + review.Rating + "|" + review.Text
	}
	return ""
}
