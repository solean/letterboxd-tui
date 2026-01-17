package letterboxd

import "testing"

func TestFollowingActivityURL(t *testing.T) {
	url := followingActivityURL("jane", "foo=bar; com.xk72.webparts.csrf=token123")
	if url == "" {
		t.Fatalf("expected URL")
	}
	if want := BaseURL + "/ajax/activity-pagination/jane/following/?"; url[:len(want)] != want {
		t.Fatalf("unexpected URL prefix: %q", url)
	}
	if got := followingActivityURL("jane", "foo=bar"); got == url {
		t.Fatalf("expected csrf to change URL")
	}
}

func TestCookieValue(t *testing.T) {
	if got := cookieValue("", "x"); got != "" {
		t.Fatalf("expected empty cookie value, got %q", got)
	}
	if got := cookieValue("a=b; c=d", "c"); got != "d" {
		t.Fatalf("expected d, got %q", got)
	}
	if got := cookieValue("csrf=123; com.xk72.webparts.csrf=abc", "com.xk72.webparts.csrf"); got != "abc" {
		t.Fatalf("expected abc, got %q", got)
	}
}

func TestCompactSpaces(t *testing.T) {
	if got := compactSpaces("  hello   world \n"); got != "hello world" {
		t.Fatalf("unexpected compactSpaces: %q", got)
	}
}
