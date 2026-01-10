package ui

import (
	"fmt"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"

	"letterboxd-tui/internal/letterboxd"
)

type diaryMsg struct {
	items []letterboxd.DiaryEntry
	err   error
}

type watchlistMsg struct {
	items []letterboxd.WatchlistItem
	err   error
}

type filmMsg struct {
	film letterboxd.Film
	err  error
}

type profileMsg struct {
	profile letterboxd.Profile
	err     error
}

type activityMsg struct {
	items []letterboxd.ActivityItem
	err   error
	tab   tab
}

type errMsg struct {
	err error
}

type openMsg struct {
	err error
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchProfileCmd(m.client, m.profileUser),
		fetchDiaryCmd(m.client, m.username),
		fetchWatchlistCmd(m.client, m.username),
		fetchActivityCmd(m.client, m.username, tabActivity),
		fetchActivityCmd(m.client, m.username, tabFollowing),
	)
}

func fetchProfileCmd(client *letterboxd.Client, username string) tea.Cmd {
	return func() tea.Msg {
		profile, err := client.Profile(username)
		return profileMsg{profile: profile, err: err}
	}
}

func fetchDiaryCmd(client *letterboxd.Client, username string) tea.Cmd {
	return func() tea.Msg {
		items, err := client.Diary(username)
		return diaryMsg{items: items, err: err}
	}
}

func fetchWatchlistCmd(client *letterboxd.Client, username string) tea.Cmd {
	return func() tea.Msg {
		items, err := client.Watchlist(username)
		return watchlistMsg{items: items, err: err}
	}
}

func fetchFilmCmd(client *letterboxd.Client, filmURL, username string) tea.Cmd {
	return func() tea.Msg {
		film, err := client.Film(filmURL, username)
		return filmMsg{film: film, err: err}
	}
}

func fetchActivityCmd(client *letterboxd.Client, username string, which tab) tea.Cmd {
	return func() tea.Msg {
		var (
			items []letterboxd.ActivityItem
			err   error
		)
		switch which {
		case tabActivity:
			items, err = client.Activity(username)
		case tabFollowing:
			items, err = client.FollowingActivity(username)
		default:
			return errMsg{err: fmt.Errorf("unknown activity tab")}
		}
		return activityMsg{tab: which, items: items, err: err}
	}
}

func openBrowserCmd(url string) tea.Cmd {
	return func() tea.Msg {
		if url == "" {
			return openMsg{err: fmt.Errorf("missing profile URL")}
		}
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}
		if err := cmd.Start(); err != nil {
			return openMsg{err: err}
		}
		return openMsg{}
	}
}
