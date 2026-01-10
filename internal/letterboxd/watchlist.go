package letterboxd

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseWatchlist(doc *goquery.Document) ([]WatchlistItem, error) {
	var items []WatchlistItem
	doc.Find(".js-watchlist-main-content .react-component").Each(func(_ int, item *goquery.Selection) {
		title := strings.TrimSpace(item.AttrOr("data-item-name", ""))
		filmURL := strings.TrimSpace(item.AttrOr("data-item-link", ""))
		if title == "" || filmURL == "" {
			return
		}
		if !strings.Contains(filmURL, "/film/") {
			return
		}
		if strings.HasPrefix(filmURL, "/") {
			filmURL = BaseURL + filmURL
		}
		year := ""
		if open := strings.LastIndex(title, "("); open != -1 {
			if close := strings.LastIndex(title, ")"); close > open {
				year = strings.TrimSpace(title[open+1 : close])
				title = strings.TrimSpace(title[:open])
			}
		}
		items = append(items, WatchlistItem{
			Title:   title,
			FilmURL: filmURL,
			Year:    year,
		})
	})
	return items, nil
}
