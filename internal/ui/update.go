package ui

import (
	tea "github.com/charmbracelet/bubbletea"

	"letterboxd-tui/internal/letterboxd"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		bodyHeight := max(1, m.height-3)
		m.viewport.Width = m.width
		m.viewport.Height = bodyHeight
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
		case "o":
			if m.profileModal {
				return m, openBrowserCmd(letterboxd.ProfileURL(m.modalUser))
			} else if m.activeTab == tabProfile {
				return m, openBrowserCmd(letterboxd.ProfileURL(m.profileUser))
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
