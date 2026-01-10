package letterboxd

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func parseActivity(doc *goquery.Document) ([]ActivityItem, error) {
	var items []ActivityItem
	doc.Find("section.activity-row").Each(func(_ int, row *goquery.Selection) {
		kind := strings.TrimSpace(row.AttrOr("class", ""))
		summarySel := row.Find(".activity-summary").First()
		summary := strings.TrimSpace(summarySel.Text())
		when := strings.TrimSpace(row.Find("time.time").First().AttrOr("datetime", ""))
		actorSel := row.Find(".activity-summary a.name").First()
		if actorSel.Length() == 0 {
			actorSel = row.Find(".attribution-detail a.owner").First()
		}
		actor := strings.TrimSpace(actorSel.Text())
		actorURL, _ := actorSel.Attr("href")
		if actorURL != "" && strings.HasPrefix(actorURL, "/") {
			actorURL = BaseURL + actorURL
		}
		targetSel := row.Find(".activity-summary a.target").First()
		title := strings.TrimSpace(targetSel.Text())
		filmURL, _ := targetSel.Attr("href")
		if title == "" {
			titleSel := row.Find("h2.name a").First()
			title = strings.TrimSpace(titleSel.Text())
			filmURL, _ = titleSel.Attr("href")
		}
		if filmURL != "" && strings.HasPrefix(filmURL, "/") {
			filmURL = BaseURL + filmURL
		}
		rating := strings.TrimSpace(row.Find(".rating").First().Text())
		parts := parseSummaryParts(summarySel)
		items = append(items, ActivityItem{
			Summary:  summary,
			When:     when,
			Title:    title,
			FilmURL:  filmURL,
			Rating:   rating,
			Kind:     kind,
			Actor:    actor,
			ActorURL: actorURL,
			Parts:    parts,
		})
	})
	return items, nil
}

func parseSummaryParts(summary *goquery.Selection) []SummaryPart {
	var parts []SummaryPart
	if summary == nil {
		return parts
	}
	summary.Contents().Each(func(_ int, node *goquery.Selection) {
		n := node.Get(0)
		if n == nil {
			return
		}
		switch n.Type {
		case html.TextNode:
			addSummaryPart(&parts, n.Data, "text")
		case html.ElementNode:
			if node.Is("a") {
				class := node.AttrOr("class", "")
				text := node.Text()
				switch {
				case strings.Contains(class, "target"):
					if n != nil {
						extracted := extractTargetTitle(n)
						if extracted != "" {
							text = extracted
						}
					}
					addSummaryPart(&parts, text, "movie")
				case strings.Contains(class, "name"):
					addSummaryPart(&parts, text, "user")
				default:
					addSummaryPart(&parts, text, "text")
				}
				return
			}
			if node.Is("span") {
				class := node.AttrOr("class", "")
				text := node.Text()
				if strings.Contains(class, "rating") {
					addSummaryPart(&parts, text, "rating")
				} else {
					addSummaryPart(&parts, text, "text")
				}
				return
			}
			if node.Is("strong") {
				class := node.AttrOr("class", "")
				text := node.Text()
				if strings.Contains(class, "name") {
					addSummaryPart(&parts, text, "user")
				} else {
					addSummaryPart(&parts, text, "text")
				}
				return
			}
			addSummaryPart(&parts, node.Text(), "text")
		}
	})
	return parts
}

func addSummaryPart(parts *[]SummaryPart, text, kind string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	*parts = append(*parts, SummaryPart{Text: text, Kind: kind})
}

func extractTargetTitle(node *html.Node) string {
	var parts []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil {
			return
		}
		switch n.Type {
		case html.TextNode:
			parts = append(parts, n.Data)
		case html.ElementNode:
			if n.Data == "span" {
				class := attrValue(n, "class")
				if strings.Contains(class, "context") || strings.Contains(class, "rating") {
					return
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
	}
	walk(node)
	return compactSpaces(strings.Join(parts, " "))
}

func attrValue(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}
