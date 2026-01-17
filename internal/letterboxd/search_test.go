package letterboxd

import "testing"

func TestParseSearchResults(t *testing.T) {
	html := `
	<ul>
		<li class="search-result">
			<div class="react-component" data-item-name="The Matrix (1999)" data-item-slug="the-matrix" data-item-link="/film/the-matrix/" data-film-id="1"></div>
		</li>
		<li class="search-result">
			<div class="film-title-wrapper">
				<a href="/film/memento/">Memento</a>
				<small><a>2000</a></small>
			</div>
		</li>
	</ul>`
	doc := docFromHTML(t, html)
	results := parseSearchResults(doc)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Title != "The Matrix" || results[0].Year != "1999" {
		t.Fatalf("unexpected first result: %+v", results[0])
	}
	if results[1].Title != "Memento" || results[1].Year != "2000" {
		t.Fatalf("unexpected second result: %+v", results[1])
	}
}

func TestSearchFilmsMissingQuery(t *testing.T) {
	client := NewClient(nil, "")
	if _, err := client.SearchFilms(" "); err == nil {
		t.Fatalf("expected error for empty query")
	}
}
