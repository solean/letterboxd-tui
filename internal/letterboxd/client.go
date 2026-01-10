package letterboxd

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Client struct {
	HTTP   *http.Client
	Cookie string
}

func NewClient(httpClient *http.Client, cookie string) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 12 * time.Second}
	}
	return &Client{
		HTTP:   httpClient,
		Cookie: strings.TrimSpace(cookie),
	}
}

func (c *Client) Profile(username string) (Profile, error) {
	url := fmt.Sprintf("%s/%s/", BaseURL, username)
	doc, err := c.fetchDocument(url)
	if err != nil {
		return Profile{}, err
	}
	return parseProfile(doc)
}

func (c *Client) Diary(username string) ([]DiaryEntry, error) {
	url := fmt.Sprintf("%s/%s/diary/", BaseURL, username)
	doc, err := c.fetchDocument(url)
	if err != nil {
		return nil, err
	}
	return parseDiary(doc)
}

func (c *Client) Watchlist(username string) ([]WatchlistItem, error) {
	url := fmt.Sprintf("%s/%s/watchlist/", BaseURL, username)
	doc, err := c.fetchDocument(url)
	if err != nil {
		return nil, err
	}
	return parseWatchlist(doc)
}

func (c *Client) Film(filmURL, username string) (Film, error) {
	doc, err := c.fetchDocument(filmURL)
	if err != nil {
		return Film{}, err
	}
	film, err := parseFilm(doc, filmURL)
	if err != nil {
		return film, err
	}
	if username != "" {
		userURL := userFilmURL(username, filmURL)
		if userURL != "" {
			userDoc, status, err := c.fetchDocumentAllowStatus(userURL)
			if err == nil && status == http.StatusOK {
				film.UserRating, film.UserStatus = parseUserFilm(userDoc)
			}
		}
	}
	return film, nil
}

func (c *Client) Activity(username string) ([]ActivityItem, error) {
	url := fmt.Sprintf("%s/ajax/activity-pagination/%s/", BaseURL, username)
	doc, err := c.fetchDocument(url)
	if err != nil {
		return nil, err
	}
	return parseActivity(doc)
}

func (c *Client) FollowingActivity(username string) ([]ActivityItem, error) {
	url := followingActivityURL(username, c.Cookie)
	doc, err := c.fetchDocument(url)
	if err != nil {
		return nil, err
	}
	return parseActivity(doc)
}

func (c *Client) fetchDocument(url string) (*goquery.Document, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "letterboxd-tui/0.1")
	if c.Cookie != "" {
		req.Header.Set("Cookie", c.Cookie)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}
	return goquery.NewDocumentFromReader(resp.Body)
}

func (c *Client) fetchDocumentAllowStatus(url string) (*goquery.Document, int, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "letterboxd-tui/0.1")
	if c.Cookie != "" {
		req.Header.Set("Cookie", c.Cookie)
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	return doc, resp.StatusCode, err
}
