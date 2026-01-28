package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/solean/letterboxd-tui/internal/config"
)

func TestResolveStartupFromEnv(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	t.Setenv("LETTERBOXD_USER", "alice")
	t.Setenv("LETTERBOXD_COOKIE", "foo=bar; com.xk72.webparts.csrf=csrf123")

	state, err := resolveStartup("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.username != "alice" {
		t.Fatalf("unexpected username: %q", state.username)
	}
	if state.cookie != "" {
		t.Fatalf("expected empty cookie, got %q", state.cookie)
	}
	if state.needUsername || !state.needCookie {
		t.Fatalf("expected cookie onboarding only; got user=%v cookie=%v", state.needUsername, state.needCookie)
	}
}

func TestResolveStartupFromConfig(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	t.Setenv("LETTERBOXD_USER", "")
	t.Setenv("LETTERBOXD_COOKIE", "")

	if err := config.Save(config.Config{Username: "jo", Cookie: "foo=bar; com.xk72.webparts.csrf=cfgcsrf"}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	state, err := resolveStartup("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.username != "jo" {
		t.Fatalf("unexpected username: %q", state.username)
	}
	if state.cookie != "foo=bar; com.xk72.webparts.csrf=cfgcsrf" {
		t.Fatalf("unexpected cookie: %q", state.cookie)
	}
}

func TestResolveStartupFromLegacyCookie(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })
	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	t.Setenv("LETTERBOXD_USER", "")
	t.Setenv("LETTERBOXD_COOKIE", "")

	if err := os.WriteFile(filepath.Join(dir, "cookie.txt"), []byte("legacy=1; com.xk72.webparts.csrf=legacycsrf\n"), 0600); err != nil {
		t.Fatalf("write legacy cookie: %v", err)
	}

	state, err := resolveStartup("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.cookie != "" {
		t.Fatalf("expected legacy cookie to be ignored, got %q", state.cookie)
	}
	if state.configDirty {
		t.Fatalf("did not expect config to be marked dirty")
	}
}

func TestResolveStartupNeedsOnboarding(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	t.Setenv("HOME", dir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(dir, "xdg"))
	t.Setenv("LETTERBOXD_USER", "")
	t.Setenv("LETTERBOXD_COOKIE", "")

	state, err := resolveStartup("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !state.needUsername || !state.needCookie {
		t.Fatalf("expected onboarding requirements; got user=%v cookie=%v", state.needUsername, state.needCookie)
	}
}
