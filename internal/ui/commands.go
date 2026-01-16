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

type reviewsMsg struct {
	reviews []letterboxd.Review
	err     error
	kind    string
}

type profileMsg struct {
	profile letterboxd.Profile
	err     error
	modal   bool
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

type logResultMsg struct {
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
		return profileMsg{profile: profile, err: err, modal: false}
	}
}

func fetchProfileModalCmd(client *letterboxd.Client, username string) tea.Cmd {
	return func() tea.Msg {
		profile, err := client.Profile(username)
		return profileMsg{profile: profile, err: err, modal: true}
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

func fetchReviewsCmd(client *letterboxd.Client, slug string, which string) tea.Cmd {
	return func() tea.Msg {
		var (
			revs []letterboxd.Review
			err  error
		)
		switch which {
		case "popular":
			revs, err = client.PopularReviews(slug)
		case "friends":
			revs, err = client.FriendReviews(slug)
		default:
			err = fmt.Errorf("unknown reviews kind")
		}
		return reviewsMsg{reviews: revs, err: err, kind: which}
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

func saveDiaryEntryCmd(client *letterboxd.Client, req letterboxd.DiaryEntryRequest) tea.Cmd {
	return func() tea.Msg {
		err := client.SaveDiaryEntry(req)
		return logResultMsg{err: err}
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
