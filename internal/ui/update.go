package ui

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
		m.refreshModalViewport()
		return m, nil
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
		if rm.kind == "popular" {
			m.popReviews = rm.reviews
			m.popReviewsErr = rm.err
		} else {
			m.friendReviews = rm.reviews
			m.friendReviewsErr = rm.err
		}
		m.refreshModalViewport()
		return m, nil
	}

	if m.logModal {
		return m.updateLogModal(msg)
	}

	switch ev := msg.(type) {
	case tea.KeyMsg:
		switch ev.String() {
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
		m.diary = ev.items
		m.diaryErr = ev.err
		m.loading = false
	case watchlistMsg:
		m.watchlist = ev.items
		m.watchErr = ev.err
		m.loading = false
	case filmMsg:
		m.film = ev.film
		m.filmErr = ev.err
		m.loading = false
		m.refreshModalViewport()
		if ev.film.Slug != "" {
			return m, tea.Batch(
				fetchReviewsCmd(m.client, ev.film.Slug, "popular"),
				fetchReviewsCmd(m.client, ev.film.Slug, "friends"),
			)
		}
	case activityMsg:
		if ev.tab == tabActivity {
			m.activity = ev.items
			m.activityErr = ev.err
		} else {
			m.following = ev.items
			m.followErr = ev.err
		}
		m.loading = false
	case errMsg:
		m.diaryErr = ev.err
		m.loading = false
	case openMsg:
		if ev.err != nil {
			m.profileErr = ev.err
		}
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
	legend := theme.subtle.Render(renderLegend(*m))
	width, height := modalDimensions(m.width, m.height)
	innerWidth := width - 4
	innerHeight := height - 2
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
