package letterboxd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Review struct {
	Author string
	Rating string
	Text   string
	Link   string
}

func (c *Client) PopularReviews(slug string, page int) ([]Review, error) {
	if slug == "" {
		return nil, c.wrapDebug(fmt.Errorf("missing slug"))
	}
	if page < 1 {
		page = 1
	}
	url := fmt.Sprintf("%s/film/%s/reviews/by/activity/page/%d/", BaseURL, slug, page)
	doc, err := c.fetchDocument(url)
	if err != nil {
		return nil, err
	}
	return parseReviews(doc)
}

func (c *Client) FriendReviews(slug, username string, page int) ([]Review, error) {
	if slug == "" {
		return nil, c.wrapDebug(fmt.Errorf("missing slug"))
	}
	if page < 1 {
		page = 1
	}
	if strings.TrimSpace(username) != "" {
		url := fmt.Sprintf("%s/%s/friends/film/%s/reviews/by/activity/", BaseURL, username, slug)
		if page > 1 {
			url = fmt.Sprintf("%s/%s/friends/film/%s/reviews/by/activity/page/%d/", BaseURL, username, slug, page)
		}
		if doc, _, err := c.fetchDocumentAllowStatus(url); err == nil {
			return parseReviews(doc)
		}
	}
	url := fmt.Sprintf("%s/csi/film/%s/friend-reviews/", BaseURL, slug)
	if page > 1 {
		url = fmt.Sprintf("%s/csi/film/%s/friend-reviews/page/%d/", BaseURL, slug, page)
	}
	url += "?esiAllowUser=true"
	doc, err := c.fetchDocument(url)
	if err != nil {
		return nil, err
	}
	return parseReviews(doc)
}

func parseReviews(doc *goquery.Document) ([]Review, error) {
	selectors := []string{
		".production-viewing",
		"[data-viewing-id], [data-review-id]",
		"li.film-detail, article.film-detail, div.film-detail",
		"article.review, li.review, div.review",
	}
	for _, selector := range selectors {
		reviews := parseReviewsBySelector(doc, selector)
		if len(reviews) > 0 {
			return reviews, nil
		}
	}
	if reviews := parseReviewsByBody(doc); len(reviews) > 0 {
		return reviews, nil
	}
	return nil, nil
}

func parseReviewsBySelector(doc *goquery.Document, selector string) []Review {
	var reviews []Review
	seen := make(map[string]struct{})
	doc.Find(selector).Each(func(_ int, view *goquery.Selection) {
		review := parseReviewFromSelection(view)
		key := reviewKey(review)
		if key == "" {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		reviews = append(reviews, review)
	})
	return reviews
}

func parseReviewsByBody(doc *goquery.Document) []Review {
	var reviews []Review
	seen := make(map[string]struct{})
	doc.Find(".js-review-body, .body-text, .review-body, .review").Each(func(_ int, body *goquery.Selection) {
		container := body.ParentsFiltered(".production-viewing, [data-viewing-id], [data-review-id], .film-detail, .review").First()
		if container.Length() == 0 {
			container = body.Parent()
		}
		review := parseReviewFromSelection(container)
		if review.Text == "" {
			review.Text = compactSpaces(body.Text())
		}
		key := reviewKey(review)
		if key == "" {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		reviews = append(reviews, review)
	})
	return reviews
}

func parseReviewFromSelection(view *goquery.Selection) Review {
	author := firstAttr(view, "[data-owner-name]", "data-owner-name")
	if author == "" {
		author = strings.TrimSpace(view.AttrOr("data-owner-name", ""))
	}
	if author == "" {
		author = strings.TrimSpace(view.AttrOr("data-username", ""))
	}
	if author == "" {
		author = firstAttr(view, "[data-username]", "data-username")
	}
	if author == "" {
		author = firstText(view,
			".displayname",
			".name a",
			".name",
			"a.name",
			".person-summary .name",
		)
	}
	if author == "" {
		author = authorFromLinks(view)
	}
	rating := ratingFromSelection(view)
	text := firstText(view,
		".js-review-body",
		".body-text",
		".review-body",
		".review",
	)
	link := firstAttr(view, ".attribution-detail a.context", "href")
	if link == "" {
		link = firstAttr(view, "a[href*='/review/']", "href")
	}
	if link != "" && strings.HasPrefix(link, "/") {
		link = BaseURL + link
	}
	return Review{
		Author: author,
		Rating: rating,
		Text:   compactSpaces(text),
		Link:   link,
	}
}

func authorFromLinks(view *goquery.Selection) string {
	var author string
	view.Find("a").EachWithBreak(func(_ int, link *goquery.Selection) bool {
		text := strings.TrimSpace(link.Text())
		if text == "" {
			return true
		}
		href := strings.TrimSpace(link.AttrOr("href", ""))
		if href == "" || !strings.HasPrefix(href, "/") {
			return true
		}
		lower := strings.ToLower(href)
		if strings.Contains(lower, "/review/") ||
			strings.Contains(lower, "/film/") ||
			strings.Contains(lower, "/lists/") ||
			strings.Contains(lower, "/list/") {
			return true
		}
		if strings.Count(strings.Trim(href, "/"), "/") > 0 {
			return true
		}
		author = text
		return false
	})
	return author
}

func ratingFromSelection(view *goquery.Selection) string {
	rating := strings.TrimSpace(view.Find(".content-reactions-strip .rating").First().Text())
	if rating != "" {
		return rating
	}
	ratingSel := view.Find(".rating").First()
	rating = strings.TrimSpace(ratingSel.Text())
	if rating != "" {
		return rating
	}
	if label := strings.TrimSpace(ratingSel.AttrOr("aria-label", "")); label != "" {
		if stars := extractStars(label); stars != "" {
			return stars
		}
	}
	if label := strings.TrimSpace(ratingSel.AttrOr("data-original-title", "")); label != "" {
		if stars := extractStars(label); stars != "" {
			return stars
		}
	}
	if value := strings.TrimSpace(ratingSel.AttrOr("data-rating", "")); value != "" {
		return ratingFromValue(value)
	}
	if value := strings.TrimSpace(ratingSel.AttrOr("data-rated", "")); value != "" {
		return ratingFromValue(value)
	}
	if value := strings.TrimSpace(ratingSel.AttrOr("data-value", "")); value != "" {
		return ratingFromValue(value)
	}
	return ""
}

func extractStars(value string) string {
	var out strings.Builder
	for _, r := range value {
		if r == '★' || r == '½' {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func ratingFromValue(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return value
	}
	if num < 0 || num > 5 {
		return value
	}
	return starsFromValue(num)
}

func starsFromValue(value float64) string {
	if value <= 0 {
		return ""
	}
	full := int(value)
	half := value-float64(full) >= 0.5
	var out strings.Builder
	for i := 0; i < full; i++ {
		out.WriteRune('★')
	}
	if half {
		out.WriteRune('½')
	}
	return out.String()
}

func firstText(view *goquery.Selection, selectors ...string) string {
	for _, selector := range selectors {
		if selector == "" {
			continue
		}
		if text := strings.TrimSpace(view.Find(selector).First().Text()); text != "" {
			return text
		}
	}
	return ""
}

func firstAttr(view *goquery.Selection, selector, attr string) string {
	if selector == "" || attr == "" {
		return ""
	}
	if sel := view.Find(selector).First(); sel.Length() > 0 {
		return strings.TrimSpace(sel.AttrOr(attr, ""))
	}
	return ""
}

func reviewKey(review Review) string {
	if review.Link != "" {
		return review.Link
	}
	if review.Author == "" && review.Text == "" && review.Rating == "" {
		return ""
	}
	return review.Author + "|" + review.Rating + "|" + review.Text
}
