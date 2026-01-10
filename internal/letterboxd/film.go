package letterboxd

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseFilm(doc *goquery.Document, url string) (Film, error) {
	var film Film
	film.URL = url
	title := strings.TrimSpace(doc.Find(`meta[property="og:title"]`).AttrOr("content", ""))
	description := strings.TrimSpace(doc.Find(`meta[property="og:description"]`).AttrOr("content", ""))
	director := strings.TrimSpace(doc.Find(`meta[name="twitter:data1"]`).AttrOr("content", ""))
	avgRating := strings.TrimSpace(doc.Find(`meta[name="twitter:data2"]`).AttrOr("content", ""))
	runtime := strings.TrimSpace(findRuntime(doc))
	cast := parseTopBilledCast(doc, 6)

	year := ""
	if open := strings.LastIndex(title, "("); open != -1 {
		if close := strings.LastIndex(title, ")"); close > open {
			year = strings.TrimSpace(title[open+1 : close])
			title = strings.TrimSpace(title[:open])
		}
	}
	film.Title = title
	film.Year = year
	film.Description = description
	film.Director = director
	film.AvgRating = avgRating
	film.Runtime = runtime
	film.Cast = cast
	return film, nil
}

func findRuntime(doc *goquery.Document) string {
	text := compactSpaces(doc.Find("p.text-link.text-footer").First().Text())
	if text == "" {
		return ""
	}
	fields := strings.Fields(text)
	for i := 0; i < len(fields); i++ {
		if strings.HasSuffix(fields[i], "mins") {
			return fields[i]
		}
		if i+1 < len(fields) && strings.HasSuffix(fields[i+1], "mins") {
			return fields[i] + " " + fields[i+1]
		}
	}
	return ""
}

func parseTopBilledCast(doc *goquery.Document, limit int) []string {
	var cast []string
	doc.Find("#tab-cast .cast-list a.text-slug").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		name := strings.TrimSpace(sel.Text())
		if name == "" {
			return true
		}
		cast = append(cast, name)
		if limit > 0 && len(cast) >= limit {
			return false
		}
		return true
	})
	return cast
}

func parseUserFilm(doc *goquery.Document) (string, string) {
	status := strings.TrimSpace(doc.Find(".film-viewing-info-wrapper .context").First().Text())
	status = strings.TrimSuffix(status, "by")
	status = strings.TrimSpace(strings.TrimSuffix(status, "By"))
	rating := strings.TrimSpace(doc.Find(".content-reactions-strip .rating").First().Text())
	return rating, status
}
