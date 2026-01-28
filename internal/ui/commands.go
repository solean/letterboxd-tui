package ui

import (
	"fmt"
	"os/exec"
	"runtime"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/solean/letterboxd-tui/internal/letterboxd"
)

var execCommand = exec.Command

type diaryMsg struct {
	items []letterboxd.DiaryEntry
	err   error
	page  int
}

type watchlistMsg struct {
	items []letterboxd.WatchlistItem
	err   error
	page  int
}

type filmMsg struct {
	film letterboxd.Film
	err  error
}

type searchMsg struct {
	results []letterboxd.SearchResult
	err     error
}

type reviewsMsg struct {
	reviews []letterboxd.Review
	err     error
	kind    string
	page    int
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
	after string
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

type watchlistResultMsg struct {
	err         error
	inWatchlist bool
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		fetchProfileCmd(m.client, m.profileUser),
		fetchDiaryCmd(m.client, m.username, 1),
		fetchWatchlistCmd(m.client, m.username, 1),
		fetchActivityCmd(m.client, m.username, tabActivity, ""),
	}
	if m.hasCookie() {
		cmds = append(cmds, fetchActivityCmd(m.client, m.username, tabFollowing, ""))
	}
	return tea.Batch(cmds...)
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

func fetchDiaryCmd(client *letterboxd.Client, username string, page int) tea.Cmd {
	return func() tea.Msg {
		items, err := client.Diary(username, page)
		return diaryMsg{items: items, err: err, page: page}
	}
}

func fetchWatchlistCmd(client *letterboxd.Client, username string, page int) tea.Cmd {
	return func() tea.Msg {
		items, err := client.Watchlist(username, page)
		return watchlistMsg{items: items, err: err, page: page}
	}
}

func fetchFilmCmd(client *letterboxd.Client, filmURL, username string) tea.Cmd {
	return func() tea.Msg {
		film, err := client.Film(filmURL, username)
		return filmMsg{film: film, err: err}
	}
}

func fetchSearchCmd(client *letterboxd.Client, query string) tea.Cmd {
	return func() tea.Msg {
		results, err := client.SearchFilms(query)
		return searchMsg{results: results, err: err}
	}
}

func fetchReviewsCmd(client *letterboxd.Client, slug, username string, which string, page int) tea.Cmd {
	return func() tea.Msg {
		var (
			revs []letterboxd.Review
			err  error
		)
		switch which {
		case "popular":
			revs, err = client.PopularReviews(slug, page)
		case "friends":
			revs, err = client.FriendReviews(slug, username, page)
		default:
			err = fmt.Errorf("unknown reviews kind")
		}
		return reviewsMsg{reviews: revs, err: err, kind: which, page: page}
	}
}

func fetchActivityCmd(client *letterboxd.Client, username string, which tab, after string) tea.Cmd {
	return func() tea.Msg {
		var (
			items []letterboxd.ActivityItem
			err   error
		)
		switch which {
		case tabActivity:
			items, err = client.Activity(username, after)
		case tabFollowing:
			items, err = client.FollowingActivity(username, after)
		default:
			return errMsg{err: fmt.Errorf("unknown activity tab")}
		}
		return activityMsg{tab: which, items: items, err: err, after: after}
	}
}

func saveDiaryEntryCmd(client *letterboxd.Client, req letterboxd.DiaryEntryRequest) tea.Cmd {
	return func() tea.Msg {
		err := client.SaveDiaryEntry(req)
		return logResultMsg{err: err}
	}
}

func setWatchlistCmd(client *letterboxd.Client, req letterboxd.WatchlistRequest, inWatchlist bool) tea.Cmd {
	return func() tea.Msg {
		err := client.SetWatchlist(req, inWatchlist)
		return watchlistResultMsg{err: err, inWatchlist: inWatchlist}
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
			cmd = execCommand("open", url)
		case "windows":
			cmd = execCommand("cmd", "/c", "start", url)
		default:
			cmd = execCommand("xdg-open", url)
		}
		if err := cmd.Start(); err != nil {
			return openMsg{err: err}
		}
		return openMsg{}
	}
}
