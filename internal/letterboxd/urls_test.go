package letterboxd

import "testing"

func TestProfileURL(t *testing.T) {
	if got := ProfileURL(""); got != "" {
		t.Fatalf("expected empty profile URL, got %q", got)
	}
	if got := ProfileURL("jane"); got != BaseURL+"/jane/" {
		t.Fatalf("unexpected profile URL: %q", got)
	}
}

func TestUsernameFromURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "relative", in: "/jdoe/", want: "jdoe"},
		{name: "absolute", in: BaseURL + "/jane/", want: "jane"},
		{name: "invalid", in: "jane", want: ""},
	}
	for _, tt := range tests {
		if got := UsernameFromURL(tt.in); got != tt.want {
			t.Fatalf("%s: expected %q, got %q", tt.name, tt.want, got)
		}
	}
}

func TestNormalizeFilmURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "slug", in: "film/inception", want: BaseURL + "/film/inception/"},
		{name: "absolute", in: BaseURL + "/film/inception/", want: BaseURL + "/film/inception/"},
		{name: "embedded", in: BaseURL + "/jane/film/inception/diary/", want: BaseURL + "/film/inception/"},
		{name: "invalid", in: "https://example.com", want: ""},
	}
	for _, tt := range tests {
		if got := NormalizeFilmURL(tt.in); got != tt.want {
			t.Fatalf("%s: expected %q, got %q", tt.name, tt.want, got)
		}
	}
}

func TestFilmSlug(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{in: "", want: ""},
		{in: BaseURL + "/film/inception/", want: "inception"},
		{in: "/film/inception/", want: "inception"},
		{in: BaseURL + "/jane/film/inception/diary/", want: "inception"},
		{in: "not a film", want: ""},
	}
	for _, tt := range tests {
		if got := FilmSlug(tt.in); got != tt.want {
			t.Fatalf("expected %q, got %q for %q", tt.want, got, tt.in)
		}
	}
}

func TestUserFilmURL(t *testing.T) {
	if got := userFilmURL("", BaseURL+"/film/inception/"); got != "" {
		t.Fatalf("expected empty URL with missing username, got %q", got)
	}
	if got := userFilmURL("jane", ""); got != "" {
		t.Fatalf("expected empty URL with missing film, got %q", got)
	}
	if got := userFilmURL("jane", BaseURL+"/film/inception/"); got != BaseURL+"/jane/film/inception/" {
		t.Fatalf("unexpected user film URL: %q", got)
	}
}

func TestDiaryURL(t *testing.T) {
	if got := diaryURL("jane", 1, DiarySortDefault); got != BaseURL+"/jane/diary/" {
		t.Fatalf("unexpected diary URL: %q", got)
	}
	if got := diaryURL("jane", 2, DiarySortDefault); got != BaseURL+"/jane/diary/films/page/2/" {
		t.Fatalf("unexpected diary URL: %q", got)
	}
	if got := diaryURL("jane", 1, DiarySortAddedEarliest); got != BaseURL+"/jane/diary/films/by/added-earliest/" {
		t.Fatalf("unexpected diary URL: %q", got)
	}
	if got := diaryURL("jane", 3, DiarySortRating); got != BaseURL+"/jane/diary/films/by/entry-rating/page/3/" {
		t.Fatalf("unexpected diary URL: %q", got)
	}
}

func TestWatchlistURL(t *testing.T) {
	if got := watchlistURL("jane", 1, WatchlistSortDefault); got != BaseURL+"/jane/watchlist/" {
		t.Fatalf("unexpected watchlist URL: %q", got)
	}
	if got := watchlistURL("jane", 2, WatchlistSortDefault); got != BaseURL+"/jane/watchlist/page/2/" {
		t.Fatalf("unexpected watchlist URL: %q", got)
	}
	if got := watchlistURL("jane", 1, WatchlistSortName); got != BaseURL+"/jane/watchlist/by/name/" {
		t.Fatalf("unexpected watchlist URL: %q", got)
	}
	if got := watchlistURL("jane", 2, WatchlistSortRating); got != BaseURL+"/jane/watchlist/by/rating/page/2/" {
		t.Fatalf("unexpected watchlist URL: %q", got)
	}
	if got := watchlistURL("jane", 4, WatchlistSortRelease); got != BaseURL+"/jane/watchlist/by/release/page/4/" {
		t.Fatalf("unexpected watchlist URL: %q", got)
	}
}
