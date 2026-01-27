package ui

import (
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"letterboxd-tui/internal/letterboxd"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newHTTPResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func newStubClient(handler func(*http.Request) (*http.Response, error)) *letterboxd.Client {
	return letterboxd.NewClient(&http.Client{Transport: roundTripFunc(handler)}, "com.xk72.webparts.csrf=csrf123")
}

func TestFetchCommands(t *testing.T) {
	profileHTML := `<div class="profile-stats"><div class="profile-statistic"><span class="value">1</span><span class="definition">Films</span></div></div>`
	diaryHTML := `<tr class="diary-entry-row"><td class="col-monthdate"><span class="month">Jan</span><span class="year">2024</span></td><td class="col-daydate"><span class="daydate">1</span></td><h2 class="name"><a href="/film/inception/">Inception</a></h2></tr>`
	watchlistHTML := `<div class="js-watchlist-main-content"><div class="react-component" data-item-name="Inception (2010)" data-item-link="/film/inception/"></div></div>`
	activityHTML := `<section class="activity-row"><div class="activity-summary"><a class="name" href="/jane/">Jane</a> watched <a class="target" href="/film/inception/">Inception</a></div><time class="time" datetime="2024-01-05T00:00:00Z"></time></section>`
	searchHTML := `<li class="search-result"><div class="react-component" data-item-name="Inception (2010)" data-item-link="/film/inception/"></div></li>`
	filmHTML := `<meta property="og:title" content="Inception (2010)"><div data-film-id="123"></div>`
	reviewsHTML := `<div class="production-viewing"><span class="displayname">Jane</span></div>`

	client := newStubClient(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/jane/":
			return newHTTPResponse(http.StatusOK, profileHTML), nil
		case "/jane/diary/":
			return newHTTPResponse(http.StatusOK, diaryHTML), nil
		case "/jane/watchlist/":
			return newHTTPResponse(http.StatusOK, watchlistHTML), nil
		case "/ajax/activity-pagination/jane/":
			return newHTTPResponse(http.StatusOK, activityHTML), nil
		case "/ajax/activity-pagination/jane/following/":
			return newHTTPResponse(http.StatusOK, activityHTML), nil
		case "/s/search/films/inception/":
			return newHTTPResponse(http.StatusOK, searchHTML), nil
		case "/film/inception/":
			return newHTTPResponse(http.StatusOK, filmHTML), nil
		case "/film/inception/json":
			return newHTTPResponse(http.StatusOK, `{"lid":"lid123","id":123}`), nil
		case "/ajax/film/inception/popular-reviews/":
			return newHTTPResponse(http.StatusOK, reviewsHTML), nil
		case "/csi/film/inception/friend-reviews/":
			return newHTTPResponse(http.StatusOK, reviewsHTML), nil
		default:
			return newHTTPResponse(http.StatusNotFound, "nope"), nil
		}
	})

	if msg := fetchProfileCmd(client, "jane")(); msg.(profileMsg).err != nil {
		t.Fatalf("unexpected profile error")
	}
	if msg := fetchDiaryCmd(client, "jane", 1)(); msg.(diaryMsg).err != nil {
		t.Fatalf("unexpected diary error")
	}
	if msg := fetchWatchlistCmd(client, "jane", 1)(); msg.(watchlistMsg).err != nil {
		t.Fatalf("unexpected watchlist error")
	}
	if msg := fetchActivityCmd(client, "jane", tabActivity, "")(); msg.(activityMsg).err != nil {
		t.Fatalf("unexpected activity error")
	}
	if msg := fetchActivityCmd(client, "jane", tabFollowing, "")(); msg.(activityMsg).err != nil {
		t.Fatalf("unexpected following error")
	}
	if msg := fetchSearchCmd(client, "inception")(); msg.(searchMsg).err != nil {
		t.Fatalf("unexpected search error")
	}
	if msg := fetchFilmCmd(client, letterboxd.BaseURL+"/film/inception/", "")(); msg.(filmMsg).err != nil {
		t.Fatalf("unexpected film error")
	}
	if msg := fetchReviewsCmd(client, "inception", "popular", 1)(); msg.(reviewsMsg).err != nil {
		t.Fatalf("unexpected reviews error")
	}
}

func TestFetchReviewsUnknownKind(t *testing.T) {
	client := newStubClient(func(req *http.Request) (*http.Response, error) {
		return newHTTPResponse(http.StatusOK, ""), nil
	})
	msg := fetchReviewsCmd(client, "slug", "nope", 1)().(reviewsMsg)
	if msg.err == nil {
		t.Fatalf("expected error for unknown kind")
	}
}

func TestFetchActivityUnknownTab(t *testing.T) {
	client := newStubClient(func(req *http.Request) (*http.Response, error) {
		return newHTTPResponse(http.StatusOK, ""), nil
	})
	if msg := fetchActivityCmd(client, "jane", tabFilm, "")(); msg.(errMsg).err == nil {
		t.Fatalf("expected error for unknown tab")
	}
}

func TestSaveDiaryEntryCmd(t *testing.T) {
	client := newStubClient(func(req *http.Request) (*http.Response, error) {
		return newHTTPResponse(http.StatusOK, "ok"), nil
	})
	req := letterboxd.DiaryEntryRequest{ViewingUID: "film:123"}
	if msg := saveDiaryEntryCmd(client, req)(); msg.(logResultMsg).err != nil {
		t.Fatalf("unexpected save diary error")
	}
}

func TestSetWatchlistCmd(t *testing.T) {
	client := newStubClient(func(req *http.Request) (*http.Response, error) {
		return newHTTPResponse(http.StatusOK, "ok"), nil
	})
	req := letterboxd.WatchlistRequest{WatchlistID: "lid123"}
	if msg := setWatchlistCmd(client, req, true)(); msg.(watchlistResultMsg).err != nil {
		t.Fatalf("unexpected watchlist error")
	}
}

func TestOpenBrowserCmd(t *testing.T) {
	if msg := openBrowserCmd("")().(openMsg); msg.err == nil {
		t.Fatalf("expected error for missing URL")
	}

	oldExec := execCommand
	defer func() { execCommand = oldExec }()

	execCommand = func(name string, args ...string) *exec.Cmd {
		if runtime.GOOS == "windows" {
			return exec.Command("cmd", "/c", "exit", "0")
		}
		return exec.Command("true")
	}

	if msg := openBrowserCmd("http://example.com")().(openMsg); msg.err != nil {
		t.Fatalf("unexpected open error: %v", msg.err)
	}
}
