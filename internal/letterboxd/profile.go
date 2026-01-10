package letterboxd

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseProfile(doc *goquery.Document) (Profile, error) {
	var profile Profile
	doc.Find(".profile-stats .profile-statistic").Each(func(_ int, stat *goquery.Selection) {
		value := strings.TrimSpace(stat.Find(".value").First().Text())
		label := strings.TrimSpace(stat.Find(".definition").First().Text())
		url, _ := stat.Find("a").First().Attr("href")
		if url != "" && strings.HasPrefix(url, "/") {
			url = BaseURL + url
		}
		if value != "" && label != "" {
			profile.Stats = append(profile.Stats, ProfileStat{
				Label: label,
				Value: value,
				URL:   url,
			})
		}
	})

	doc.Find("#favourites .posteritem .react-component").Each(func(_ int, fav *goquery.Selection) {
		title := strings.TrimSpace(fav.AttrOr("data-item-name", ""))
		filmURL := strings.TrimSpace(fav.AttrOr("data-item-link", ""))
		if filmURL != "" && strings.HasPrefix(filmURL, "/") {
			filmURL = BaseURL + filmURL
		}
		year := ""
		if open := strings.LastIndex(title, "("); open != -1 {
			if close := strings.LastIndex(title, ")"); close > open {
				year = strings.TrimSpace(title[open+1 : close])
				title = strings.TrimSpace(title[:open])
			}
		}
		if title != "" {
			profile.Favorites = append(profile.Favorites, FavoriteFilm{
				Title:   title,
				FilmURL: filmURL,
				Year:    year,
			})
		}
	})

	doc.Find("section.timeline .activity-summary").Each(func(_ int, summary *goquery.Selection) {
		line := compactSpaces(summary.Text())
		if line == "" {
			return
		}
		if !strings.Contains(strings.ToLower(line), "watched") {
			return
		}
		profile.Recent = append(profile.Recent, line)
	})
	return profile, nil
}
