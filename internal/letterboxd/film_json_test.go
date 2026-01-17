package letterboxd

import (
	"fmt"
	"net/http"
	"testing"
)

func TestFilmJSONMissingSlug(t *testing.T) {
	client := NewClient(nil, "")
	if _, err := client.filmJSON(" "); err == nil {
		t.Fatalf("expected error for missing slug")
	}
}

func TestFilmJSONSuccess(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/film/inception/json" {
			return newHTTPResponse(http.StatusNotFound, "nope", nil), nil
		}
		body := `{"lid":"lid123","uid":"uid123","id":42,"slug":"inception","url":"/film/inception/","inWatchlist":true}`
		return newHTTPResponse(http.StatusOK, body, map[string]string{"Content-Type": "application/json"}), nil
	})
	payload, err := client.filmJSON("inception")
	if err != nil {
		t.Fatalf("filmJSON error: %v", err)
	}
	if payload.ID != 42 || payload.LID != "lid123" || payload.UID != "uid123" {
		t.Fatalf("unexpected payload: %+v", payload)
	}
	if payload.InWatchlist == nil || !*payload.InWatchlist {
		t.Fatalf("expected inWatchlist true")
	}
}

func TestFilmJSONStatusError(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		return newHTTPResponse(http.StatusBadRequest, "bad", nil), nil
	})
	if _, err := client.filmJSON("inception"); err == nil {
		t.Fatalf("expected error")
	} else if got := fmt.Sprint(err); got == "" {
		t.Fatalf("expected error message")
	}
}
