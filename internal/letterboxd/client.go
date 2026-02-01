package letterboxd

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/solean/letterboxd-tui/internal/version"
	"golang.org/x/net/http2"
)

type Client struct {
	HTTP   *http.Client
	Cookie string
	Debug  bool

	fallbackHTTP   *http.Client
	forceHTTP2HTTP *http.Client
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

func (c *Client) Diary(username string, page int, sort DiarySort) ([]DiaryEntry, error) {
	url := diaryURL(username, page, sort)
	doc, err := c.fetchDocument(url)
	if err != nil {
		return nil, c.wrapDebug(err)
	}
	entries, err := parseDiary(doc)
	return entries, c.wrapDebug(err)
}

func (c *Client) Watchlist(username string, page int, sort WatchlistSort) ([]WatchlistItem, error) {
	url := watchlistURL(username, page, sort)
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
	if headers == nil {
		headers = map[string]string{}
	}
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
	req.Header.Set("User-Agent", resolvedUserAgent())
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
}

func (c *Client) fetchDocumentStatus(url string, headers map[string]string) (*goquery.Document, int, error) {
	const maxAttempts = 2
	useFallback := false
	forceHTTP2 := false
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
		client := c.HTTP
		if forceHTTP2 {
			client = c.forceHTTP2Client()
		} else if useFallback {
			client = c.fallbackClient()
		}
		resp, err := client.Do(req)
		if err != nil {
			if !forceHTTP2 && isHTTP2PrefaceError(err) {
				forceHTTP2 = true
				continue
			}
			return nil, 0, c.wrapDebug(err)
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			doc, err := goquery.NewDocumentFromReader(resp.Body)
			resp.Body.Close()
			return doc, resp.StatusCode, c.wrapDebug(err)
		}
		body := readBodySnippet(resp.Body)
		resp.Body.Close()
		isChallenge := isCloudflareChallenge(resp.StatusCode, body)
		if isChallenge && attempt < maxAttempts {
			if !useFallback {
				useFallback = true
				continue
			}
			time.Sleep(cloudflareBackoff(attempt))
			continue
		}
		if attempt < maxAttempts && shouldRetryStatus(resp.StatusCode) {
			if !useFallback {
				useFallback = true
				continue
			}
			time.Sleep(cloudflareBackoff(attempt))
			continue
		}
		if isChallenge {
			return nil, resp.StatusCode, c.cloudflareError(req, resp, body)
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

func shouldRetryStatus(status int) bool {
	switch status {
	case http.StatusForbidden, http.StatusTooManyRequests, http.StatusServiceUnavailable:
		return true
	default:
		return false
	}
}

func cloudflareBackoff(attempt int) time.Duration {
	base := 300 * time.Millisecond
	if attempt <= 0 {
		return base
	}
	return base * time.Duration(1<<attempt)
}

func resolvedUserAgent() string {
	if ua := strings.TrimSpace(os.Getenv("LETTERBOXD_USER_AGENT")); ua != "" {
		return ua
	}
	appUA := version.UserAgent()
	if appUA == "" {
		return defaultUserAgent
	}
	return fmt.Sprintf("%s %s", defaultUserAgent, appUA)
}

func (c *Client) fallbackClient() *http.Client {
	if c.fallbackHTTP != nil {
		return c.fallbackHTTP
	}
	base := c.HTTP
	if base == nil {
		base = &http.Client{Timeout: 12 * time.Second}
	}
	transport := base.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	baseTransport, ok := transport.(*http.Transport)
	if !ok {
		c.fallbackHTTP = base
		return c.fallbackHTTP
	}
	transportClone := baseTransport.Clone()
	transportClone.ForceAttemptHTTP2 = false
	transportClone.TLSNextProto = map[string]func(string, *tls.Conn) http.RoundTripper{}
	c.fallbackHTTP = &http.Client{
		Timeout:       base.Timeout,
		Transport:     transportClone,
		CheckRedirect: base.CheckRedirect,
		Jar:           base.Jar,
	}
	return c.fallbackHTTP
}

func (c *Client) forceHTTP2Client() *http.Client {
	if c.forceHTTP2HTTP != nil {
		return c.forceHTTP2HTTP
	}
	base := c.HTTP
	if base == nil {
		base = &http.Client{Timeout: 12 * time.Second}
	}
	transport := base.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	baseTransport, ok := transport.(*http.Transport)
	if !ok {
		c.forceHTTP2HTTP = base
		return c.forceHTTP2HTTP
	}
	transportClone := baseTransport.Clone()
	transportClone.ForceAttemptHTTP2 = true
	if err := http2.ConfigureTransport(transportClone); err != nil {
		// If configuration fails, still keep ForceAttemptHTTP2 enabled.
	}
	c.forceHTTP2HTTP = &http.Client{
		Timeout:       base.Timeout,
		Transport:     transportClone,
		CheckRedirect: base.CheckRedirect,
		Jar:           base.Jar,
	}
	return c.forceHTTP2HTTP
}

func cloneTransport(rt http.RoundTripper) *http.Transport {
	if rt == nil {
		return http.DefaultTransport.(*http.Transport).Clone()
	}
	if transport, ok := rt.(*http.Transport); ok {
		return transport.Clone()
	}
	return http.DefaultTransport.(*http.Transport).Clone()
}

func (c *Client) cloudflareError(req *http.Request, resp *http.Response, body string) error {
	hint := fmt.Sprintf("Cloudflare challenge detected (status %d). Refresh your Letterboxd cookie from a browser (include com.xk72.webparts.csrf and cf_clearance) and try again.", resp.StatusCode)
	if !c.Debug {
		return fmt.Errorf("%s", hint)
	}
	detail := c.httpStatusErrorWithBody(req, resp, body)
	return fmt.Errorf("%s\n%w", hint, detail)
}

func isHTTP2PrefaceError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "malformed HTTP response") ||
		strings.Contains(msg, "HTTP/1.x transport connection broken")
}
