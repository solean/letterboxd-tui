package letterboxd

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestSaveDiaryEntryMissingUID(t *testing.T) {
	client := NewClient(nil, "com.xk72.webparts.csrf=csrf123")
	if err := client.SaveDiaryEntry(DiaryEntryRequest{}); err == nil {
		t.Fatalf("expected error for missing viewing UID")
	}
}

func TestSaveDiaryEntryMissingCSRF(t *testing.T) {
	client := NewClient(nil, "")
	req := DiaryEntryRequest{ViewingUID: "film:123"}
	if err := client.SaveDiaryEntry(req); err == nil {
		t.Fatalf("expected error for missing csrf")
	}
}

func TestSaveDiaryEntrySuccess(t *testing.T) {
	var captured url.Values
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		if req.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", req.Method)
		}
		body, _ := io.ReadAll(req.Body)
		captured, _ = url.ParseQuery(string(body))
		return newHTTPResponse(http.StatusOK, "ok", nil), nil
	})
	req := DiaryEntryRequest{
		ViewingUID:       "film:123",
		WatchedDate:      "2024-01-01",
		RatingValue:      11,
		Review:           "Great",
		ContainsSpoilers: true,
		Rewatch:          true,
		Tags:             "tag1,tag2",
		Liked:            true,
		Privacy:          "Friends",
		Draft:            true,
		Referer:          BaseURL + "/film/inception/",
		JSONResponse:     true,
	}
	if err := client.SaveDiaryEntry(req); err != nil {
		t.Fatalf("SaveDiaryEntry error: %v", err)
	}
	if got := captured.Get("viewingableUid"); got != "film:123" {
		t.Fatalf("unexpected viewingableUid: %q", got)
	}
	if got := captured.Get("rating"); got != "10" {
		t.Fatalf("expected rating clamp to 10, got %q", got)
	}
	if captured.Get("rewatch") != "true" || captured.Get("containsSpoilers") != "true" {
		t.Fatalf("expected rewatch/spoilers true")
	}
	if got := strings.TrimSpace(captured.Get("privacyPolicyStr")); got != "Friends" {
		t.Fatalf("unexpected privacy: %q", got)
	}
}

func TestSaveDiaryEntryStatusError(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		return newHTTPResponse(http.StatusInternalServerError, "oops", nil), nil
	})
	req := DiaryEntryRequest{
		ViewingUID:   "film:123",
		WatchedDate:  "2024-01-01",
		JSONResponse: true,
	}
	if err := client.SaveDiaryEntry(req); err == nil {
		t.Fatalf("expected error")
	}
}

func TestSaveDiaryEntryJSONError(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		return newHTTPResponse(http.StatusOK, `{"error":"nope"}`, map[string]string{"Content-Type": "application/json"}), nil
	})
	req := DiaryEntryRequest{
		ViewingUID:   "film:123",
		WatchedDate:  "2024-01-01",
		JSONResponse: true,
	}
	if err := client.SaveDiaryEntry(req); err == nil || !strings.Contains(err.Error(), "nope") {
		t.Fatalf("expected json error, got %v", err)
	}
}

func TestSaveDiaryEntryJSONResultFalse(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		return newHTTPResponse(http.StatusOK, `{"result":false}`, map[string]string{"Content-Type": "application/json"}), nil
	})
	req := DiaryEntryRequest{
		ViewingUID:   "film:123",
		WatchedDate:  "2024-01-01",
		JSONResponse: true,
	}
	if err := client.SaveDiaryEntry(req); err == nil || !strings.Contains(err.Error(), "unknown error") {
		t.Fatalf("expected unknown error for false result, got %v", err)
	}
}
