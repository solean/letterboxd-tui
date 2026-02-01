package letterboxd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
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

	const maxAttempts = 1
	useFallback := false
	reqURL := fmt.Sprintf("%s/s/save-diary-entry", BaseURL)
	encoded := values.Encode()
	for attempt := 0; attempt <= maxAttempts; attempt++ {
		httpReq, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(encoded))
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

		client := c.HTTP
		if useFallback {
			client = c.fallbackClient()
		}
		resp, err := client.Do(httpReq)
		if err != nil {
			return c.wrapDebug(err)
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if req.JSONResponse {
				body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
				resp.Body.Close()
				if err != nil {
					return c.wrapDebug(err)
				}
				if errMsg := diarySaveError(body); errMsg != "" {
					return c.wrapDebug(fmt.Errorf("save diary entry failed: %s", errMsg))
				}
				return nil
			}
			resp.Body.Close()
			return nil
		}
		snippet := ""
		if data, _ := io.ReadAll(io.LimitReader(resp.Body, 512)); len(data) > 0 {
			snippet = strings.TrimSpace(string(data))
		}
		resp.Body.Close()
		isChallenge := isCloudflareChallenge(resp.StatusCode, snippet)
		if attempt < maxAttempts && (isChallenge || shouldRetryStatus(resp.StatusCode)) {
			if !useFallback {
				useFallback = true
				continue
			}
			time.Sleep(cloudflareBackoff(attempt))
			continue
		}
		if isChallenge {
			return c.cloudflareError(httpReq, resp, snippet)
		}
		if snippet != "" {
			return c.wrapDebug(fmt.Errorf("save diary entry failed: status %d body=%q", resp.StatusCode, snippet))
		}
		return c.wrapDebug(fmt.Errorf("save diary entry failed: status %d", resp.StatusCode))
	}
	return c.wrapDebug(errors.New("save diary entry failed: retry attempts exhausted"))
}

func diarySaveError(body []byte) string {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 || !json.Valid(trimmed) {
		return ""
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(trimmed, &payload); err != nil {
		return ""
	}
	if errMsg := extractJSONError(payload); errMsg != "" {
		return errMsg
	}
	if success, ok := payload["success"].(bool); ok && !success {
		if msg := extractJSONMessage(payload["message"]); msg != "" {
			return msg
		}
		if msg := extractJSONMessage(payload["messages"]); msg != "" {
			return msg
		}
		return "unknown error"
	}
	if result, ok := payload["result"].(bool); ok {
		if !result {
			if msg := extractJSONMessage(payload["message"]); msg != "" {
				return msg
			}
			if msg := extractJSONMessage(payload["messages"]); msg != "" {
				return msg
			}
			return "unknown error"
		}
		return ""
	}
	result := strings.ToLower(extractJSONString(payload, "result"))
	if result != "" && result != "success" && result != "ok" && result != "true" {
		if msg := extractJSONMessage(payload["message"]); msg != "" {
			return msg
		}
		if msg := extractJSONMessage(payload["messages"]); msg != "" {
			return msg
		}
		if result == "false" {
			return "unknown error"
		}
		return result
	}
	return ""
}

func extractJSONError(payload map[string]interface{}) string {
	if msg := extractJSONMessage(payload["error"]); msg != "" {
		return msg
	}
	if msg := extractJSONMessage(payload["errors"]); msg != "" {
		return msg
	}
	return ""
}

func extractJSONString(payload map[string]interface{}, key string) string {
	val, ok := payload[key]
	if !ok {
		return ""
	}
	switch typed := val.(type) {
	case string:
		return strings.TrimSpace(typed)
	case fmt.Stringer:
		return strings.TrimSpace(typed.String())
	case float64, int, int64:
		return strings.TrimSpace(fmt.Sprint(typed))
	default:
		return ""
	}
}

func extractJSONMessage(val interface{}) string {
	msgs := extractJSONMessages(val)
	if len(msgs) == 0 {
		return ""
	}
	return strings.Join(msgs, "; ")
}

func extractJSONMessages(val interface{}) []string {
	switch typed := val.(type) {
	case nil:
		return nil
	case string:
		if trimmed := strings.TrimSpace(typed); trimmed != "" {
			return []string{trimmed}
		}
		return nil
	case bool:
		if typed {
			return []string{"unknown error"}
		}
		return nil
	case []string:
		return normalizeStrings(typed)
	case []interface{}:
		var msgs []string
		for _, item := range typed {
			msgs = append(msgs, extractJSONMessages(item)...)
		}
		return normalizeStrings(msgs)
	case map[string]interface{}:
		if msg := extractJSONString(typed, "message"); msg != "" {
			return []string{msg}
		}
		if msg := extractJSONString(typed, "error"); msg != "" {
			return []string{msg}
		}
		if msg := extractJSONString(typed, "detail"); msg != "" {
			return []string{msg}
		}
		if msg := extractJSONString(typed, "title"); msg != "" {
			return []string{msg}
		}
		var msgs []string
		for _, item := range typed {
			msgs = append(msgs, extractJSONMessages(item)...)
		}
		return normalizeStrings(msgs)
	default:
		return nil
	}
}

func normalizeStrings(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(items))
	normalized := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	return normalized
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
