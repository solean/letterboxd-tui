package letterboxd

import "testing"

func TestParseFilm(t *testing.T) {
	html := `
		<html><head>
		<meta property="og:title" content="Inception (2010)">
		<meta property="og:description" content="Dreams inside dreams.">
		<meta name="twitter:data1" content="Christopher Nolan">
		<meta name="twitter:data2" content="4.2">
		</head>
		<body>
			<p class="text-link text-footer">148 mins</p>
			<div id="tab-cast">
				<div class="cast-list">
					<a class="text-slug">Leonardo DiCaprio</a>
					<a class="text-slug">Elliot Page</a>
				</div>
			</div>
			<div data-film-id="12345"></div>
		</body></html>`
	doc := docFromHTML(t, html)
	film, err := parseFilm(doc, BaseURL+"/film/inception/")
	if err != nil {
		t.Fatalf("parseFilm error: %v", err)
	}
	if film.Title != "Inception" || film.Year != "2010" {
		t.Fatalf("unexpected title/year: %+v", film)
	}
	if film.Director != "Christopher Nolan" || film.AvgRating != "4.2" {
		t.Fatalf("unexpected director/rating: %+v", film)
	}
	if film.Runtime != "148 mins" {
		t.Fatalf("unexpected runtime: %q", film.Runtime)
	}
	if film.FilmID != "12345" || film.ViewingUID != "film:12345" {
		t.Fatalf("unexpected film id: %+v", film)
	}
	if len(film.Cast) != 2 {
		t.Fatalf("unexpected cast: %+v", film.Cast)
	}
}

func TestFindRuntime(t *testing.T) {
	doc := docFromHTML(t, `<p class="text-link text-footer">2 hrs 5 mins</p>`)
	if got := findRuntime(doc); got != "5 mins" {
		t.Fatalf("unexpected runtime: %q", got)
	}
}

func TestParseTopBilledCastLimit(t *testing.T) {
	doc := docFromHTML(t, `<div id="tab-cast"><div class="cast-list">
		<a class="text-slug">A</a><a class="text-slug">B</a><a class="text-slug">C</a>
	</div></div>`)
	cast := parseTopBilledCast(doc, 2)
	if len(cast) != 2 || cast[0] != "A" || cast[1] != "B" {
		t.Fatalf("unexpected cast: %+v", cast)
	}
}

func TestFindFilmIDFromScript(t *testing.T) {
	doc := docFromHTML(t, `<script>var viewingable.uid = "film:98765";</script>`)
	if got := findFilmID(doc); got != "98765" {
		t.Fatalf("expected 98765, got %q", got)
	}
}

func TestIsDigits(t *testing.T) {
	if !isDigits("123") {
		t.Fatalf("expected digits")
	}
	if isDigits("12a") || isDigits("") {
		t.Fatalf("expected non-digits to fail")
	}
}

func TestParseUserFilm(t *testing.T) {
	doc := docFromHTML(t, `
		<div class="film-viewing-info-wrapper"><span class="context">Watched by</span></div>
		<div class="content-reactions-strip"><span class="rating">★★★★</span></div>`)
	rating, status := parseUserFilm(doc)
	if rating != "★★★★" {
		t.Fatalf("unexpected rating: %q", rating)
	}
	if status != "Watched" {
		t.Fatalf("unexpected status: %q", status)
	}
}
