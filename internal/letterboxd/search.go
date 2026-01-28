package letterboxd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func (c *Client) SearchFilms(query string) ([]SearchResult, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, c.wrapDebug(fmt.Errorf("missing query"))
	}
	escaped := url.PathEscape(query)
	endpoint := fmt.Sprintf("%s/s/search/films/%s/", BaseURL, escaped)
	doc, err := c.fetchDocument(endpoint)
	if err != nil {
		return nil, err
	}
	return parseSearchResults(doc), nil
}

func parseSearchResults(doc *goquery.Document) []SearchResult {
	var results []SearchResult
	doc.Find("li.search-result").Each(func(_ int, item *goquery.Selection) {
		comp := item.Find(".react-component").First()
		title := strings.TrimSpace(comp.AttrOr("data-item-name", ""))
		slug := strings.TrimSpace(comp.AttrOr("data-item-slug", ""))
		filmURL := strings.TrimSpace(comp.AttrOr("data-item-link", ""))
		filmID := strings.TrimSpace(comp.AttrOr("data-film-id", ""))

		if filmURL == "" {
			filmURL = strings.TrimSpace(item.Find(".film-title-wrapper a").First().AttrOr("href", ""))
		}
		if filmURL != "" && strings.HasPrefix(filmURL, "/") {
			filmURL = BaseURL + filmURL
		}

		year := strings.TrimSpace(item.Find(".film-title-wrapper small a").First().Text())
		if title == "" {
			title = strings.TrimSpace(item.Find(".film-title-wrapper a").First().Text())
		}
		if year == "" {
			if open := strings.LastIndex(title, "("); open != -1 {
				if close := strings.LastIndex(title, ")"); close > open {
					year = strings.TrimSpace(title[open+1 : close])
					title = strings.TrimSpace(title[:open])
				}
			}
		}
		if title == "" || filmURL == "" {
			return
		}
		results = append(results, SearchResult{
			Title:   title,
			Year:    year,
			FilmURL: filmURL,
			Slug:    slug,
			FilmID:  filmID,
		})
	})
	return results
}
