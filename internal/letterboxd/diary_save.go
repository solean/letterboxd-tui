package letterboxd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type DiaryEntryRequest struct {
	ViewingUID       string
	WatchedDate      string
	RatingValue      int // 0-10 where 10 == 5 stars
	Review           string
	ContainsSpoilers bool
	Rewatch          bool
	Tags             string
	Liked            bool
	Privacy          string // "", "Anyone", "Friends", "You"
	Draft            bool
	Referer          string
	JSONResponse     bool
}

func (c *Client) SaveDiaryEntry(req DiaryEntryRequest) error {
	if req.ViewingUID == "" {
		return c.wrapDebug(errors.New("missing viewing UID"))
	}
	csrf := cookieValue(c.Cookie, "com.xk72.webparts.csrf")
	if csrf == "" {
		return c.wrapDebug(errors.New("missing csrf token in cookie"))
	}
	values := url.Values{}
	if req.JSONResponse {
		values.Set("json", "true")
	}
	values.Set("__csrf", csrf)
	values.Set("viewingId", "")
	values.Set("viewingableUid", req.ViewingUID)
	values.Set("viewingableUID", req.ViewingUID) // observed in browser requests
	values.Set("rating", strconv.Itoa(clamp(req.RatingValue, 0, 10)))
	if req.Rewatch {
		values.Set("rewatch", "true")
	}
	if req.WatchedDate != "" {
		values.Set("specifiedDate", "true")
		values.Set("viewingDateStr", req.WatchedDate)
	}
	if req.Review != "" {
		values.Set("review", req.Review)
	}
	if req.ContainsSpoilers {
		values.Set("containsSpoilers", "true")
	}
	if strings.TrimSpace(req.Tags) != "" {
		values.Set("tags", req.Tags)
	}
	if req.Liked {
		values.Set("liked", "true")
	}
	if req.Privacy != "" {
		values.Set("privacyPolicyStr", req.Privacy)
	}
	if req.Draft {
		values.Set("privacyPolicyDraft", "true")
	}

	reqURL := fmt.Sprintf("%s/s/save-diary-entry", BaseURL)
	httpReq, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(values.Encode()))
	if err != nil {
		return c.wrapDebug(err)
	}
	applyDefaultHeaders(httpReq)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Origin", BaseURL)
	httpReq.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	httpReq.Header.Set("X-Requested-With", "XMLHttpRequest")
	if strings.TrimSpace(req.Referer) != "" {
		httpReq.Header.Set("Referer", req.Referer)
	}
	if c.Cookie != "" {
		httpReq.Header.Set("Cookie", c.Cookie)
	}

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return c.wrapDebug(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := ""
		if data, _ := io.ReadAll(io.LimitReader(resp.Body, 512)); len(data) > 0 {
			snippet = strings.TrimSpace(string(data))
		}
		if snippet != "" {
			return c.wrapDebug(fmt.Errorf("save diary entry failed: status %d body=%q", resp.StatusCode, snippet))
		}
		return c.wrapDebug(fmt.Errorf("save diary entry failed: status %d", resp.StatusCode))
	}
	return nil
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
