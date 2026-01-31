package letterboxd

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type WatchlistRequest struct {
	WatchlistID  string
	FilmID       string
	FilmSlug     string
	Referer      string
	JSONResponse bool
}

func (c *Client) AddToWatchlist(req WatchlistRequest) error {
	return c.SetWatchlist(req, true)
}

func (c *Client) RemoveFromWatchlist(req WatchlistRequest) error {
	return c.SetWatchlist(req, false)
}

func (c *Client) SetWatchlist(req WatchlistRequest, inWatchlist bool) error {
	watchlistID := strings.TrimSpace(req.WatchlistID)
	slug := strings.TrimSpace(req.FilmSlug)
	filmID := strings.TrimSpace(req.FilmID)
	if watchlistID == "" && slug == "" && filmID == "" {
		return c.wrapDebug(errors.New("missing film id"))
	}
	csrf := cookieValue(c.Cookie, "com.xk72.webparts.csrf")
	if csrf == "" {
		return c.wrapDebug(errors.New("missing csrf token in cookie"))
	}

	var lastErr error
	if watchlistID == "" && slug != "" {
		if meta, err := c.filmJSON(slug); err == nil && meta.LID != "" {
			watchlistID = meta.LID
		}
	}
	if watchlistID != "" {
		err := c.patchWatchlist(fmt.Sprintf("%s/api/v0/me/watchlist/%s", BaseURL, watchlistID), csrf, req.Referer, inWatchlist)
		if err == nil {
			return nil
		}
		lastErr = c.wrapDebug(err)
	}

	values := url.Values{}
	if req.JSONResponse {
		values.Set("json", "true")
	}
	values.Set("__csrf", csrf)
	if filmID != "" {
		values.Set("filmId", filmID)
		values.Set("filmID", filmID)
	}

	if slug != "" {
		status, err := c.postWatchlist(fmt.Sprintf("%s/film/%s/watchlist/", BaseURL, slug), values, req.Referer)
		if err == nil {
			return nil
		}
		lastErr = c.wrapDebug(err)
		if status != http.StatusNotFound && status != http.StatusMethodNotAllowed {
			return c.wrapDebug(err)
		}
	}
	if filmID != "" {
		_, err := c.postWatchlist(fmt.Sprintf("%s/ajax/film/%s/watchlist/", BaseURL, filmID), values, req.Referer)
		if err == nil {
			return nil
		}
		lastErr = c.wrapDebug(err)
	}
	if lastErr != nil {
		return lastErr
	}
	return c.wrapDebug(errors.New("unable to update watchlist"))
}

func (c *Client) patchWatchlist(reqURL, csrf, referer string, inWatchlist bool) error {
	payload := fmt.Sprintf(`{"inWatchlist":%t}`, inWatchlist)
	httpReq, err := http.NewRequest(http.MethodPatch, reqURL, strings.NewReader(payload))
	if err != nil {
		return c.wrapDebug(err)
	}
	applyDefaultHeaders(httpReq)
	httpReq.Header.Set("Content-Type", "application/json; charset=UTF-8")
	httpReq.Header.Set("Accept", "*/*")
	httpReq.Header.Set("Origin", BaseURL)
	httpReq.Header.Set("X-CSRF-Token", csrf)
	if strings.TrimSpace(referer) != "" {
		httpReq.Header.Set("Referer", referer)
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
		if isCloudflareChallenge(resp.StatusCode, snippet) {
			return c.cloudflareError(httpReq, resp, snippet)
		}
		if snippet != "" {
			return c.wrapDebug(fmt.Errorf("watchlist update failed: status %d body=%q", resp.StatusCode, snippet))
		}
		return c.wrapDebug(fmt.Errorf("watchlist update failed: status %d", resp.StatusCode))
	}
	return nil
}

func (c *Client) postWatchlist(reqURL string, values url.Values, referer string) (int, error) {
	httpReq, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(values.Encode()))
	if err != nil {
		return 0, c.wrapDebug(err)
	}
	applyDefaultHeaders(httpReq)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Origin", BaseURL)
	httpReq.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	httpReq.Header.Set("X-Requested-With", "XMLHttpRequest")
	if strings.TrimSpace(referer) != "" {
		httpReq.Header.Set("Referer", referer)
	}
	if c.Cookie != "" {
		httpReq.Header.Set("Cookie", c.Cookie)
	}

	resp, err := c.HTTP.Do(httpReq)
	if err != nil {
		return 0, c.wrapDebug(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		snippet := ""
		if data, _ := io.ReadAll(io.LimitReader(resp.Body, 512)); len(data) > 0 {
			snippet = strings.TrimSpace(string(data))
		}
		if isCloudflareChallenge(resp.StatusCode, snippet) {
			return resp.StatusCode, c.cloudflareError(httpReq, resp, snippet)
		}
		if snippet != "" {
			return resp.StatusCode, c.wrapDebug(fmt.Errorf("add to watchlist failed: status %d body=%q", resp.StatusCode, snippet))
		}
		return resp.StatusCode, c.wrapDebug(fmt.Errorf("add to watchlist failed: status %d", resp.StatusCode))
	}
	return resp.StatusCode, nil
}
