package letterboxd

import (
	"net/http"
	"net/url"
	"testing"
)

func TestSetWatchlistMissingFilmID(t *testing.T) {
	client := NewClient(nil, "com.xk72.webparts.csrf=csrf123")
	if err := client.SetWatchlist(WatchlistRequest{}, true); err == nil {
		t.Fatalf("expected error for missing film id")
	}
}

func TestSetWatchlistMissingCSRF(t *testing.T) {
	client := NewClient(nil, "")
	req := WatchlistRequest{WatchlistID: "id123"}
	if err := client.SetWatchlist(req, true); err == nil {
		t.Fatalf("expected error for missing csrf")
	}
}

func TestSetWatchlistPatchSuccess(t *testing.T) {
	var gotMethod, gotPath string
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		gotMethod = req.Method
		gotPath = req.URL.Path
		return newHTTPResponse(http.StatusOK, "ok", nil), nil
	})
	req := WatchlistRequest{WatchlistID: "lid123", Referer: BaseURL + "/film/inception/"}
	if err := client.SetWatchlist(req, true); err != nil {
		t.Fatalf("SetWatchlist error: %v", err)
	}
	if gotMethod != http.MethodPatch || gotPath != "/api/v0/me/watchlist/lid123" {
		t.Fatalf("unexpected request: %s %s", gotMethod, gotPath)
	}
}

func TestSetWatchlistFallbackToFilmID(t *testing.T) {
	var seen []string
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		seen = append(seen, req.Method+" "+req.URL.Path)
		switch req.URL.Path {
		case "/film/inception/json":
			return newHTTPResponse(http.StatusOK, `{"lid":""}`, nil), nil
		case "/film/inception/watchlist/":
			return newHTTPResponse(http.StatusNotFound, "no", nil), nil
		case "/ajax/film/123/watchlist/":
			return newHTTPResponse(http.StatusOK, "ok", nil), nil
		default:
			return newHTTPResponse(http.StatusNotFound, "no", nil), nil
		}
	})
	req := WatchlistRequest{FilmSlug: "inception", FilmID: "123", JSONResponse: true}
	if err := client.SetWatchlist(req, true); err != nil {
		t.Fatalf("SetWatchlist error: %v", err)
	}
	if len(seen) < 3 {
		t.Fatalf("expected multiple requests, got %v", seen)
	}
}

func TestPatchWatchlistError(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		return newHTTPResponse(http.StatusBadRequest, "nope", nil), nil
	})
	err := client.patchWatchlist(BaseURL+"/api/v0/me/watchlist/lid123", "csrf123", "", true)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestPostWatchlistError(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		return newHTTPResponse(http.StatusBadRequest, "nope", nil), nil
	})
	_, err := client.postWatchlist(BaseURL+"/film/inception/watchlist/", make(url.Values), "")
	if err == nil {
		t.Fatalf("expected error")
	}
}
