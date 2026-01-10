package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"letterboxd-tui/internal/letterboxd"
	"letterboxd-tui/internal/ui"
)

func main() {
	username := flag.String("user", "cschnabel", "Letterboxd username")
	flag.Parse()

	cookie, _ := loadCookie()
	client := letterboxd.NewClient(nil, cookie)

	m := ui.NewModel(*username, client)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadCookie() (string, error) {
	if env := strings.TrimSpace(os.Getenv("LETTERBOXD_COOKIE")); env != "" {
		return env, nil
	}
	path := filepath.Join(".", "cookie.txt")
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
