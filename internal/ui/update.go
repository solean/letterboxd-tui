package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"

	"letterboxd-tui/internal/letterboxd"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = ws.Width
		m.height = ws.Height
		bodyHeight := max(1, m.height-3)
		m.viewport.Width = m.width
		m.viewport.Height = bodyHeight
		if m.logModal {
			m.logForm.setSize(m.width)
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.logSpinner, cmd = m.logSpinner.Update(msg)
		if m.logModal {
			return m, cmd
		}
		return m, nil
	}

	if m.logModal {
		return m.updateLogModal(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.modalOpen() {
				if m.activeTab == tabFilm {
					m.activeTab = m.filmReturn
					m.resetTabPosition()
				} else if m.profileModal {
					m.profileModal = false
				}
				return m, nil
			}
			return m, tea.Quit
		case "tab", "right":
			if m.activeTab == tabFilm {
				m.activeTab = m.filmReturn
			} else {
				m.activeTab = nextTab(m.activeTab)
			}
			m.resetTabPosition()
		case "left", "shift+tab":
			if m.activeTab == tabFilm {
				m.activeTab = m.filmReturn
			} else {
				m.activeTab = prevTab(m.activeTab)
			}
			m.resetTabPosition()
		case "j", "down":
			if m.modalOpen() {
				m.modalVP.LineDown(1)
			} else if m.activeTab == tabProfile {
				m.viewport.LineDown(1)
			} else {
				m.moveSelection(1)
				m.syncViewportToSelection()
			}
		case "k", "up":
			if m.modalOpen() {
				m.modalVP.LineUp(1)
			} else if m.activeTab == tabProfile {
				m.viewport.LineUp(1)
			} else {
				m.moveSelection(-1)
				m.syncViewportToSelection()
			}
		case "pgdown":
			if m.modalOpen() {
				m.modalVP.ViewDown()
			} else if m.activeTab == tabProfile {
				m.viewport.ViewDown()
			} else {
				m.pageSelection(1)
				m.syncViewportToSelection()
			}
		case "pgup":
			if m.modalOpen() {
				m.modalVP.ViewUp()
			} else if m.activeTab == tabProfile {
				m.viewport.ViewUp()
			} else {
				m.pageSelection(-1)
				m.syncViewportToSelection()
			}
		case "r":
			m.loading = true
			return m, tea.Batch(
				fetchProfileCmd(m.client, m.profileUser),
				fetchDiaryCmd(m.client, m.username),
				fetchWatchlistCmd(m.client, m.username),
				fetchActivityCmd(m.client, m.username, tabActivity),
				fetchActivityCmd(m.client, m.username, tabFollowing),
			)
		case "enter":
			if m.activeTab == tabFollowing {
				m = m.openSelectedProfile()
				return m, fetchProfileModalCmd(m.client, m.modalUser)
			} else if m.activeTab == tabDiary || m.activeTab == tabWatchlist || m.activeTab == tabActivity {
				m = m.openSelectedFilm()
				if m.activeTab == tabFilm {
					return m, fetchFilmCmd(m.client, m.film.URL, m.username)
				}
			}
		case "b":
			if m.profileModal {
				m.profileModal = false
			} else if m.activeTab == tabProfile {
				m = m.goBackProfile()
				if m.activeTab == tabProfile {
					return m, fetchProfileCmd(m.client, m.profileUser)
				}
			}
		case "l":
			if m.activeTab == tabFilm {
				m = m.startLogModal()
				return m, m.logSpinner.Tick
			}
		case "o":
			if m.profileModal {
				return m, openBrowserCmd(letterboxd.ProfileURL(m.modalUser))
			} else if m.activeTab == tabProfile {
				return m, openBrowserCmd(letterboxd.ProfileURL(m.profileUser))
			} else if m.activeTab == tabFilm {
				return m, openBrowserCmd(m.film.URL)
			}
		case "esc":
			if m.activeTab == tabFilm {
				m.activeTab = m.filmReturn
				m.resetTabPosition()
			} else if m.profileModal {
				m.profileModal = false
			}
		}
	case profileMsg:
		if msg.modal {
			m.modalProfile = msg.profile
			m.modalProfileErr = msg.err
			m.modalLoading = false
		} else {
			m.profile = msg.profile
			m.profileErr = msg.err
			m.loading = false
		}
	case diaryMsg:
		m.diary = msg.items
		m.diaryErr = msg.err
		m.loading = false
	case watchlistMsg:
		m.watchlist = msg.items
		m.watchErr = msg.err
		m.loading = false
	case filmMsg:
		m.film = msg.film
		m.filmErr = msg.err
		m.loading = false
	case activityMsg:
		if msg.tab == tabActivity {
			m.activity = msg.items
			m.activityErr = msg.err
		} else {
			m.following = msg.items
			m.followErr = msg.err
		}
		m.loading = false
	case errMsg:
		m.diaryErr = msg.err
		m.loading = false
	case openMsg:
		if msg.err != nil {
			m.profileErr = msg.err
		}
	}
	return m, nil
}

func (m Model) updateLogModal(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.KeyMsg:
		switch typed.String() {
		case "esc", "q":
			m.logModal = false
			m.logForm.submitting = false
			return m, nil
		case "tab", "right":
			m.logForm.focusField(m.logForm.focus + 1)
			return m, nil
		case "shift+tab", "left":
			m.logForm.focusField(m.logForm.focus - 1)
			return m, nil
		case "enter", "ctrl+s":
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
		case " ":
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
