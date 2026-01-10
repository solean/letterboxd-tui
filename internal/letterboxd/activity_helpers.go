package letterboxd

import (
	"fmt"
	"strings"
)

func followingActivityURL(username, cookie string) string {
	csrf := cookieValue(cookie, "com.xk72.webparts.csrf")
	query := "diaryEntries=true&reviews=true&lists=true&stories=true&reviewComments=true&listComments=true&storyComments=true&watchlistAdditions=true&reviewLikes=true&listLikes=true&storyLikes=true&follows=true&yourActivity=true&incomingActivity=true"
	if csrf != "" {
		query = query + "&__csrf=" + csrf
	}
	return fmt.Sprintf("%s/ajax/activity-pagination/%s/following/?%s", BaseURL, username, query)
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
