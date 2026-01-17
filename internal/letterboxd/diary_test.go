package letterboxd

import "testing"

func TestParseDiary(t *testing.T) {
	html := `
	<table>
		<tr class="diary-entry-row">
			<td class="col-monthdate"><span class="month">Jan</span><span class="year">2024</span></td>
			<td class="col-daydate"><span class="daydate">5</span></td>
			<td><h2 class="name"><a href="/film/inception/">Inception</a></h2></td>
			<td class="col-rating"><span class="rating">★★★★</span></td>
			<td class="js-td-rewatch icon-status-on"></td>
			<td class="js-td-review"><a href="/review"></a></td>
		</tr>
		<tr class="diary-entry-row">
			<td class="col-daydate"><span class="daydate">6</span></td>
			<td><h2 class="name"><a href="/film/memento/">Memento</a></h2></td>
		</tr>
	</table>`
	doc := docFromHTML(t, html)
	entries, err := parseDiary(doc)
	if err != nil {
		t.Fatalf("parseDiary error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Date != "Jan 5 2024" || entries[0].FilmURL != BaseURL+"/film/inception/" {
		t.Fatalf("unexpected first entry: %+v", entries[0])
	}
	if !entries[0].Rewatch || !entries[0].Review {
		t.Fatalf("expected rewatch/review true: %+v", entries[0])
	}
	if entries[1].Date != "Jan 6 2024" || entries[1].Title != "Memento" {
		t.Fatalf("unexpected second entry: %+v", entries[1])
	}
}
