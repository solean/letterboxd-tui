package letterboxd

import "testing"

func TestParseProfile(t *testing.T) {
	html := `
	<div class="profile-stats">
		<div class="profile-statistic">
			<span class="value">42</span>
			<span class="definition">Films</span>
			<a href="/jane/films/"></a>
		</div>
	</div>
	<div id="favourites">
		<div class="posteritem">
			<div class="react-component" data-item-name="Inception (2010)" data-item-link="/film/inception/"></div>
		</div>
	</div>
	<section class="timeline">
		<div class="activity-summary">Jane watched Inception</div>
		<div class="activity-summary">Jane liked a list</div>
	</section>`
	doc := docFromHTML(t, html)
	profile, err := parseProfile(doc)
	if err != nil {
		t.Fatalf("parseProfile error: %v", err)
	}
	if len(profile.Stats) != 1 || profile.Stats[0].Label != "Films" {
		t.Fatalf("unexpected stats: %+v", profile.Stats)
	}
	if len(profile.Favorites) != 1 || profile.Favorites[0].Title != "Inception" {
		t.Fatalf("unexpected favorites: %+v", profile.Favorites)
	}
	if len(profile.Recent) != 1 {
		t.Fatalf("unexpected recent: %+v", profile.Recent)
	}
}
