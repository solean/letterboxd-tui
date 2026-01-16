package letterboxd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type filmJSONResponse struct {
	LID         string `json:"lid"`
	UID         string `json:"uid"`
	ID          int    `json:"id"`
	Slug        string `json:"slug"`
	URL         string `json:"url"`
	InWatchlist *bool  `json:"inWatchlist"`
}

func (c *Client) filmJSON(slug string) (filmJSONResponse, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return filmJSONResponse{}, fmt.Errorf("missing film slug")
	}
	reqURL := fmt.Sprintf("%s/film/%s/json", BaseURL, slug)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return filmJSONResponse{}, err
	}
	req.Header.Set("User-Agent", "letterboxd-tui/0.1")
	if c.Cookie != "" {
		req.Header.Set("Cookie", c.Cookie)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return filmJSONResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return filmJSONResponse{}, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, reqURL)
	}
	var payload filmJSONResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return filmJSONResponse{}, err
	}
	return payload, nil
}
