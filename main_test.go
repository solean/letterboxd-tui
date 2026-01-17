package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCookieFromEnv(t *testing.T) {
	oldEnv := os.Getenv("LETTERBOXD_COOKIE")
	t.Setenv("LETTERBOXD_COOKIE", "cookie123")
	defer os.Setenv("LETTERBOXD_COOKIE", oldEnv)

	cookie, err := loadCookie()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cookie != "cookie123" {
		t.Fatalf("unexpected cookie: %q", cookie)
	}
}

func TestLoadCookieFromFile(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Setenv("LETTERBOXD_COOKIE", "")
	defer os.Chdir(oldWD)

	path := filepath.Join(dir, "cookie.txt")
	if err := os.WriteFile(path, []byte("filecookie\n"), 0600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	cookie, err := loadCookie()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cookie != "filecookie" {
		t.Fatalf("unexpected cookie: %q", cookie)
	}
}

func TestLoadCookieMissingFile(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Setenv("LETTERBOXD_COOKIE", "")
	defer os.Chdir(oldWD)

	if _, err := loadCookie(); err == nil {
		t.Fatalf("expected error")
	}
}
