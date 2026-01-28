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
		return filmJSONResponse{}, c.wrapDebug(fmt.Errorf("missing film slug"))
	}
	reqURL := fmt.Sprintf("%s/film/%s/json", BaseURL, slug)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return filmJSONResponse{}, c.wrapDebug(err)
	}
	applyDefaultHeaders(req)
	if c.Cookie != "" {
		req.Header.Set("Cookie", c.Cookie)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return filmJSONResponse{}, c.wrapDebug(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return filmJSONResponse{}, c.httpStatusError(req, resp)
	}
	var payload filmJSONResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return filmJSONResponse{}, c.wrapDebug(err)
	}
	return payload, nil
}
