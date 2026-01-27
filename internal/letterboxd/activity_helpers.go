package letterboxd

import (
	"fmt"
	"net/url"
	"strings"
)

func activityURL(username, after string) string {
	base := fmt.Sprintf("%s/ajax/activity-pagination/%s/", BaseURL, username)
	return withAfterParam(base, after)
}

func followingActivityURL(username, cookie, after string) string {
	csrf := cookieValue(cookie, "com.xk72.webparts.csrf")
	query := "diaryEntries=true&reviews=true&lists=true&stories=true&reviewComments=true&listComments=true&storyComments=true&watchlistAdditions=true&reviewLikes=true&listLikes=true&storyLikes=true&follows=true&yourActivity=true&incomingActivity=true"
	if csrf != "" {
		query = query + "&__csrf=" + csrf
	}
	base := fmt.Sprintf("%s/ajax/activity-pagination/%s/following/?%s", BaseURL, username, query)
	return withAfterParam(base, after)
}

func withAfterParam(rawURL, after string) string {
	if after == "" {
		return rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	query := parsed.Query()
	query.Set("after", after)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func cookieValue(cookie, key string) string {
	parts := strings.Split(cookie, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, key+"=") {
			return strings.TrimPrefix(part, key+"=")
		}
	}
	return ""
}

func compactSpaces(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}
