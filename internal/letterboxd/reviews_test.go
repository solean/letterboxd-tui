package letterboxd

import "testing"

func TestParseReviews(t *testing.T) {
	html := `
	<div class="production-viewing">
		<span class="displayname">Jane</span>
		<div class="content-reactions-strip"><span class="rating">★★★½</span></div>
		<div class="js-review-body">Great movie.</div>
		<div class="attribution-detail"><a class="context" href="/review/1/">Review</a></div>
	</div>`
	doc := docFromHTML(t, html)
	reviews, err := parseReviews(doc)
	if err != nil {
		t.Fatalf("parseReviews error: %v", err)
	}
	if len(reviews) != 1 {
		t.Fatalf("expected 1 review, got %d", len(reviews))
	}
	if reviews[0].Author != "Jane" || reviews[0].Rating != "★★★½" {
		t.Fatalf("unexpected review: %+v", reviews[0])
	}
	if reviews[0].Link != BaseURL+"/review/1/" {
		t.Fatalf("unexpected link: %q", reviews[0].Link)
	}
}
