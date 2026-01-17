package letterboxd

import (
	"net/http"
	"testing"
)

func TestNewClientDefaults(t *testing.T) {
	client := NewClient(nil, " cookie ")
	if client.HTTP == nil {
		t.Fatalf("expected HTTP client")
	}
	if client.Cookie != "cookie" {
		t.Fatalf("expected trimmed cookie, got %q", client.Cookie)
	}
}

func TestFetchDocumentStatusError(t *testing.T) {
	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		return newHTTPResponse(http.StatusTeapot, "nope", nil), nil
	})
	if _, err := client.fetchDocument(BaseURL + "/nope"); err == nil {
		t.Fatalf("expected error for non-2xx")
	}
	if _, status, err := client.fetchDocumentAllowStatus(BaseURL + "/nope"); err == nil || status != http.StatusTeapot {
		t.Fatalf("expected status error with teapot, got %v %d", err, status)
	}
}

func TestClientMethods(t *testing.T) {
	profileHTML := `<div class="profile-stats"><div class="profile-statistic">
		<span class="value">1</span><span class="definition">Films</span><a href="/jane/films/"></a>
	</div></div>`
	diaryHTML := `<tr class="diary-entry-row"><td class="col-monthdate"><span class="month">Jan</span><span class="year">2024</span></td><td class="col-daydate"><span class="daydate">1</span></td><h2 class="name"><a href="/film/inception/">Inception</a></h2></tr>`
	watchlistHTML := `<div class="js-watchlist-main-content"><div class="react-component" data-item-name="Inception (2010)" data-item-link="/film/inception/"></div></div>`
	activityHTML := `<section class="activity-row"><div class="activity-summary"><a class="name" href="/jane/">Jane</a> watched <a class="target" href="/film/inception/">Inception</a></div><time class="time" datetime="2024-01-05T00:00:00Z"></time></section>`
	searchHTML := `<li class="search-result"><div class="react-component" data-item-name="Inception (2010)" data-item-link="/film/inception/" data-item-slug="inception"></div></li>`
	filmHTML := `<meta property="og:title" content="Inception (2010)"><div data-film-id="123"></div>`
	userFilmHTML := `<div class="film-viewing-info-wrapper"><span class="context">Watched by</span></div><div class="content-reactions-strip"><span class="rating">★★★★</span></div>`
	reviewsHTML := `<div class="production-viewing"><span class="displayname">Jane</span></div>`

	client := newTestClient(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/jane/":
			return newHTTPResponse(http.StatusOK, profileHTML, nil), nil
		case "/jane/diary/":
			return newHTTPResponse(http.StatusOK, diaryHTML, nil), nil
		case "/jane/watchlist/":
			return newHTTPResponse(http.StatusOK, watchlistHTML, nil), nil
		case "/ajax/activity-pagination/jane/":
			return newHTTPResponse(http.StatusOK, activityHTML, nil), nil
		case "/ajax/activity-pagination/jane/following/":
			return newHTTPResponse(http.StatusOK, activityHTML, nil), nil
		case "/s/search/films/inception/":
			return newHTTPResponse(http.StatusOK, searchHTML, nil), nil
		case "/film/inception/":
			return newHTTPResponse(http.StatusOK, filmHTML, nil), nil
		case "/film/inception/json":
			return newHTTPResponse(http.StatusOK, `{"lid":"lid123","uid":"uid123","id":123,"url":"/film/inception/","inWatchlist":true}`, nil), nil
		case "/jane/film/inception/":
			return newHTTPResponse(http.StatusOK, userFilmHTML, nil), nil
		case "/ajax/film/inception/popular-reviews/":
			return newHTTPResponse(http.StatusOK, reviewsHTML, nil), nil
		case "/csi/film/inception/friend-reviews/":
			return newHTTPResponse(http.StatusOK, reviewsHTML, nil), nil
		default:
			return newHTTPResponse(http.StatusNotFound, "nope", nil), nil
		}
	})

	if _, err := client.Profile("jane"); err != nil {
		t.Fatalf("Profile error: %v", err)
	}
	if _, err := client.Diary("jane"); err != nil {
		t.Fatalf("Diary error: %v", err)
	}
	if _, err := client.Watchlist("jane"); err != nil {
		t.Fatalf("Watchlist error: %v", err)
	}
	if _, err := client.Activity("jane"); err != nil {
		t.Fatalf("Activity error: %v", err)
	}
	if _, err := client.FollowingActivity("jane"); err != nil {
		t.Fatalf("FollowingActivity error: %v", err)
	}
	if _, err := client.SearchFilms("inception"); err != nil {
		t.Fatalf("SearchFilms error: %v", err)
	}
	film, err := client.Film(BaseURL+"/film/inception/", "jane")
	if err != nil {
		t.Fatalf("Film error: %v", err)
	}
	if film.WatchlistID != "lid123" || film.FilmID != "123" || film.ViewingUID != "film:123" {
		t.Fatalf("unexpected film metadata: %+v", film)
	}
	if _, err := client.PopularReviews("inception"); err != nil {
		t.Fatalf("PopularReviews error: %v", err)
	}
	if _, err := client.FriendReviews("inception"); err != nil {
		t.Fatalf("FriendReviews error: %v", err)
	}
}
