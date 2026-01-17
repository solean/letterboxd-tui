package letterboxd

import "testing"

func TestParseWatchlist(t *testing.T) {
	html := `
	<div class="js-watchlist-main-content">
		<div class="react-component" data-item-name="Inception (2010)" data-item-link="/film/inception/"></div>
		<div class="react-component" data-item-name="Ignore" data-item-link="/list/123/"></div>
		<div class="react-component" data-item-name="" data-item-link="/film/empty/"></div>
	</div>`
	doc := docFromHTML(t, html)
	items, err := parseWatchlist(doc)
	if err != nil {
		t.Fatalf("parseWatchlist error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].Title != "Inception" || items[0].Year != "2010" {
		t.Fatalf("unexpected item: %+v", items[0])
	}
	if items[0].FilmURL != BaseURL+"/film/inception/" {
		t.Fatalf("unexpected film URL: %q", items[0].FilmURL)
	}
}
