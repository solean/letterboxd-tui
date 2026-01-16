package letterboxd

import (
	"fmt"
	"strings"
)

func ProfileURL(username string) string {
	if username == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/", BaseURL, username)
}

func UsernameFromURL(url string) string {
	if url == "" {
		return ""
	}
	if strings.HasPrefix(url, BaseURL) {
		url = strings.TrimPrefix(url, BaseURL)
	}
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "/") {
		return ""
	}
	parts := strings.Split(strings.Trim(url, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func NormalizeFilmURL(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return ""
	}
	if strings.HasPrefix(url, BaseURL) {
		url = strings.TrimPrefix(url, BaseURL)
	}
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	parts := strings.Split(strings.Trim(url, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	if parts[0] == "film" && len(parts) >= 2 {
		return BaseURL + "/film/" + parts[1] + "/"
	}
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == "film" && i+1 < len(parts) {
			return BaseURL + "/film/" + parts[i+1] + "/"
		}
	}
	return ""
}

func FilmSlug(url string) string {
	return filmSlug(url)
}

func userFilmURL(username, filmURL string) string {
	slug := filmSlug(filmURL)
	if slug == "" || username == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/film/%s/", BaseURL, username, slug)
}

func filmSlug(filmURL string) string {
	filmURL = NormalizeFilmURL(filmURL)
	if filmURL == "" {
		return ""
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(filmURL, BaseURL), "/"), "/")
	if len(parts) >= 2 && parts[0] == "film" {
		return parts[1]
	}
	return ""
}
