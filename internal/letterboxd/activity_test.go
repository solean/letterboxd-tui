package letterboxd

import (
	"strings"
	"testing"
)

func TestParseActivity(t *testing.T) {
	html := `
	<section class="activity-row -new">
		<div class="activity-summary">
			<a class="name" href="/jane/">Jane</a>
			<span>watched</span>
			<a class="target" href="/film/inception/">Inception</a>
			<span class="rating">★★★★</span>
		</div>
		<time class="time" datetime="2024-01-05T00:00:00Z"></time>
	</section>`
	doc := docFromHTML(t, html)
	items, err := parseActivity(doc)
	if err != nil {
		t.Fatalf("parseActivity error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	item := items[0]
	if item.Actor != "Jane" || item.Title != "Inception" {
		t.Fatalf("unexpected item: %+v", item)
	}
	if !strings.Contains(item.Kind, "activity-row") {
		t.Fatalf("expected kind to include class, got %q", item.Kind)
	}
	if len(item.Parts) == 0 {
		t.Fatalf("expected summary parts")
	}
}

func TestParseSummaryParts(t *testing.T) {
	html := `<div class="activity-summary">
		<a class="name" href="/jane/">Jane</a>
		<span>watched</span>
		<a class="target" href="/film/inception/"><span class="context">watched</span> Inception</a>
		<span class="rating">★★★★</span>
	</div>`
	doc := docFromHTML(t, html)
	sel := doc.Find(".activity-summary").First()
	parts := parseSummaryParts(sel)
	if len(parts) == 0 {
		t.Fatalf("expected parts")
	}
	foundMovie := false
	for _, part := range parts {
		if part.Kind == "movie" && strings.Contains(part.Text, "Inception") {
			foundMovie = true
		}
	}
	if !foundMovie {
		t.Fatalf("expected movie part, got %+v", parts)
	}
}

func TestExtractTargetTitle(t *testing.T) {
	html := `<a class="target" href="/film/inception/"><span class="context">watched</span><span class="rating">★★★★</span> Inception</a>`
	doc := docFromHTML(t, html)
	sel := doc.Find("a").First()
	title := ""
	if node := sel.Get(0); node != nil {
		title = extractTargetTitle(node)
	}
	if title != "Inception" {
		t.Fatalf("unexpected title: %q", title)
	}
}

func TestAttrValue(t *testing.T) {
	html := `<a class="target" href="/film/inception/">Inception</a>`
	doc := docFromHTML(t, html)
	node := doc.Find("a").First().Get(0)
	if node == nil {
		t.Fatalf("expected node")
	}
	if got := attrValue(node, "class"); got != "target" {
		t.Fatalf("unexpected attr value: %q", got)
	}
}
