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

func TestParseReviewsFilmDetail(t *testing.T) {
	html := `
	<ul>
		<li class="film-detail" data-review-id="99">
			<div class="film-detail-content">
				<a class="name" href="/jane/">Jane</a>
				<span class="rating" data-rating="4.5"></span>
				<div class="body-text">Loved it.</div>
				<a href="/review/99/">Review</a>
			</div>
		</li>
	</ul>`
	doc := docFromHTML(t, html)
	reviews, err := parseReviews(doc)
	if err != nil {
		t.Fatalf("parseReviews error: %v", err)
	}
	if len(reviews) != 1 {
		t.Fatalf("expected 1 review, got %d", len(reviews))
	}
	if reviews[0].Author != "Jane" {
		t.Fatalf("unexpected author: %q", reviews[0].Author)
	}
	if reviews[0].Rating != "★★★★½" {
		t.Fatalf("unexpected rating: %q", reviews[0].Rating)
	}
	if reviews[0].Text != "Loved it." {
		t.Fatalf("unexpected text: %q", reviews[0].Text)
	}
	if reviews[0].Link != BaseURL+"/review/99/" {
		t.Fatalf("unexpected link: %q", reviews[0].Link)
	}
}
