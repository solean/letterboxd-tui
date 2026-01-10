package letterboxd

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func parseDiary(doc *goquery.Document) ([]DiaryEntry, error) {
	var entries []DiaryEntry
	currentMonth := ""
	currentYear := ""

	doc.Find("tr.diary-entry-row").Each(func(_ int, row *goquery.Selection) {
		month := strings.TrimSpace(row.Find(".col-monthdate .month").First().Text())
		year := strings.TrimSpace(row.Find(".col-monthdate .year").First().Text())
		day := strings.TrimSpace(row.Find(".col-daydate .daydate").First().Text())
		if month != "" {
			currentMonth = month
		}
		if year != "" {
			currentYear = year
		}

		titleSel := row.Find("h2.name a").First()
		title := strings.TrimSpace(titleSel.Text())
		filmURL, _ := titleSel.Attr("href")
		if filmURL != "" && strings.HasPrefix(filmURL, "/") {
			filmURL = BaseURL + filmURL
		}
		rating := strings.TrimSpace(row.Find(".col-rating .rating").First().Text())
		rewatch := strings.Contains(row.Find(".js-td-rewatch").AttrOr("class", ""), "icon-status-on")
		review := row.Find(".js-td-review a").Length() > 0

		date := ""
		if day != "" && currentMonth != "" && currentYear != "" {
			date = fmt.Sprintf("%s %s %s", currentMonth, day, currentYear)
		}
		if title != "" {
			entries = append(entries, DiaryEntry{
				Date:    date,
				Title:   title,
				FilmURL: filmURL,
				Rating:  rating,
				Rewatch: rewatch,
				Review:  review,
			})
		}
	})
	return entries, nil
}
