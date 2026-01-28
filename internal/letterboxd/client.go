package letterboxd

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Client struct {
	HTTP   *http.Client
	Cookie string
	Debug  bool
}

const defaultUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"

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
		return Profile{}, c.wrapDebug(err)
	}
	profile, err := parseProfile(doc)
	return profile, c.wrapDebug(err)
}

func (c *Client) Diary(username string, page int) ([]DiaryEntry, error) {
	url := fmt.Sprintf("%s/%s/diary/", BaseURL, username)
	if page > 1 {
		url = fmt.Sprintf("%s/%s/diary/films/page/%d/", BaseURL, username, page)
	}
	doc, err := c.fetchDocument(url)
	if err != nil {
		return nil, c.wrapDebug(err)
	}
	entries, err := parseDiary(doc)
	return entries, c.wrapDebug(err)
}

func (c *Client) Watchlist(username string, page int) ([]WatchlistItem, error) {
	url := fmt.Sprintf("%s/%s/watchlist/", BaseURL, username)
	if page > 1 {
		url = fmt.Sprintf("%s/%s/watchlist/page/%d/", BaseURL, username, page)
	}
	doc, err := c.fetchDocument(url)
	if err != nil {
		return nil, c.wrapDebug(err)
	}
	items, err := parseWatchlist(doc)
	return items, c.wrapDebug(err)
}

func (c *Client) Film(filmURL, username string) (Film, error) {
	doc, err := c.fetchDocument(filmURL)
	if err != nil {
		return Film{}, c.wrapDebug(err)
	}
	film, err := parseFilm(doc, filmURL)
	if err != nil {
		return film, c.wrapDebug(err)
	}
	if film.Slug != "" {
		if meta, err := c.filmJSON(film.Slug); err == nil {
			if film.WatchlistID == "" && meta.LID != "" {
				film.WatchlistID = meta.LID
			}
			if film.FilmID == "" && meta.ID != 0 {
				film.FilmID = strconv.Itoa(meta.ID)
			}
			if film.ViewingUID == "" && meta.UID != "" {
				film.ViewingUID = meta.UID
			}
			if film.URL == "" && meta.URL != "" {
				film.URL = BaseURL + meta.URL
			}
			if meta.InWatchlist != nil {
				film.InWatchlist = *meta.InWatchlist
				film.WatchlistOK = true
			}
		}
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

func (c *Client) Activity(username, after string) ([]ActivityItem, error) {
	url := activityURL(username, after)
	doc, err := c.fetchDocumentWithHeaders(url, activityHeaders(username, false))
	if err != nil {
		return nil, c.wrapDebug(err)
	}
	items, err := parseActivity(doc)
	return items, c.wrapDebug(err)
}

func (c *Client) FollowingActivity(username, after string) ([]ActivityItem, error) {
	url := followingActivityURL(username, c.Cookie, after)
	doc, err := c.fetchDocumentWithHeaders(url, activityHeaders(username, true))
	if err != nil {
		return nil, c.wrapDebug(err)
	}
	items, err := parseActivity(doc)
	return items, c.wrapDebug(err)
}

func (c *Client) fetchDocument(url string) (*goquery.Document, error) {
	doc, _, err := c.fetchDocumentStatus(url, nil)
	return doc, err
}

func (c *Client) fetchDocumentWithHeaders(url string, headers map[string]string) (*goquery.Document, error) {
	doc, _, err := c.fetchDocumentStatus(url, headers)
	return doc, err
}

func activityHeaders(username string, following bool) map[string]string {
	path := fmt.Sprintf("%s/%s/activity/", BaseURL, username)
	if following {
		path = fmt.Sprintf("%s/%s/activity/following/", BaseURL, username)
	}
	return map[string]string{
		"X-Requested-With": "XMLHttpRequest",
		"Referer":          path,
	}
}

func (c *Client) fetchDocumentAllowStatus(url string) (*goquery.Document, int, error) {
	return c.fetchDocumentStatus(url, nil)
}

func applyDefaultHeaders(req *http.Request) {
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
}

func (c *Client) fetchDocumentStatus(url string, headers map[string]string) (*goquery.Document, int, error) {
	const maxAttempts = 2
	for attempt := 0; attempt <= maxAttempts; attempt++ {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, 0, c.wrapDebug(err)
		}
		applyDefaultHeaders(req)
		if c.Cookie != "" {
			req.Header.Set("Cookie", c.Cookie)
		}
		for key, val := range headers {
			if key == "" || val == "" {
				continue
			}
			req.Header.Set(key, val)
		}
		resp, err := c.HTTP.Do(req)
		if err != nil {
			return nil, 0, c.wrapDebug(err)
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			doc, err := goquery.NewDocumentFromReader(resp.Body)
			resp.Body.Close()
			return doc, resp.StatusCode, c.wrapDebug(err)
		}
		body := readBodySnippet(resp.Body)
		resp.Body.Close()
		if isCloudflareChallenge(resp.StatusCode, body) && attempt < maxAttempts {
			time.Sleep(cloudflareBackoff(attempt))
			continue
		}
		return nil, resp.StatusCode, c.httpStatusErrorWithBody(req, resp, body)
	}
	return nil, 0, c.wrapDebug(fmt.Errorf("unexpected request failure for %s", url))
}

func isCloudflareChallenge(status int, body string) bool {
	if status != http.StatusForbidden && status != http.StatusTooManyRequests && status != http.StatusServiceUnavailable {
		return false
	}
	lower := strings.ToLower(body)
	return strings.Contains(lower, "just a moment") ||
		strings.Contains(lower, "cf-chl") ||
		strings.Contains(lower, "attention required") ||
		strings.Contains(lower, "cloudflare")
}

func cloudflareBackoff(attempt int) time.Duration {
	base := 300 * time.Millisecond
	if attempt <= 0 {
		return base
	}
	return base * time.Duration(1<<attempt)
}
