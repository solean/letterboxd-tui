package letterboxd

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Review struct {
	Author string
	Rating string
	Text   string
	Link   string
}

func (c *Client) PopularReviews(slug string) ([]Review, error) {
	if slug == "" {
		return nil, fmt.Errorf("missing slug")
	}
	url := fmt.Sprintf("%s/ajax/film/%s/popular-reviews/", BaseURL, slug)
	doc, err := c.fetchDocument(url)
	if err != nil {
		return nil, err
	}
	return parseReviews(doc)
}

func (c *Client) FriendReviews(slug string) ([]Review, error) {
	if slug == "" {
		return nil, fmt.Errorf("missing slug")
	}
	url := fmt.Sprintf("%s/csi/film/%s/friend-reviews/?esiAllowUser=true", BaseURL, slug)
	doc, err := c.fetchDocument(url)
	if err != nil {
		return nil, err
	}
	return parseReviews(doc)
}

func parseReviews(doc *goquery.Document) ([]Review, error) {
	var reviews []Review
	doc.Find(".production-viewing").Each(func(_ int, view *goquery.Selection) {
		author := strings.TrimSpace(view.Find(".displayname").First().Text())
		rating := strings.TrimSpace(view.Find(".content-reactions-strip .rating").First().Text())
		text := strings.TrimSpace(view.Find(".js-review-body").First().Text())
		link := strings.TrimSpace(view.Find(".attribution-detail a.context").First().AttrOr("href", ""))
		if link != "" && strings.HasPrefix(link, "/") {
			link = BaseURL + link
		}
		if author == "" && text == "" && rating == "" {
			return
		}
		reviews = append(reviews, Review{
			Author: author,
			Rating: rating,
			Text:   compactSpaces(text),
			Link:   link,
		})
	})
	return reviews, nil
}
