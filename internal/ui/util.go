package ui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func truncate(s string, width int) string {
	if width <= 0 {
		return s
	}
	if lipgloss.Width(s) <= width {
		return s
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Left, s[:max(0, width-1)]+"…")
}

func compactSpaces(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func wrapText(s string, width int) string {
	if width <= 0 {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	var lines []string
	line := words[0]
	for _, word := range words[1:] {
		if lipgloss.Width(line)+1+lipgloss.Width(word) > width {
			lines = append(lines, line)
			line = word
		} else {
			line += " " + word
		}
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}

func appendWithSpacing(out *strings.Builder, text string) {
	if out.Len() == 0 {
		out.WriteString(text)
		return
	}
	prev := out.String()
	last := prev[len(prev)-1]
	if last != ' ' {
		out.WriteString(" ")
	}
	out.WriteString(text)
}

func modalDimensions(w, h int) (int, int) {
	width := max(50, min(96, w-6))
	height := max(10, min(24, h-6))
	return width, height
}

func formatWhen(when string) string {
	if when == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339Nano, when)
	if err != nil {
		t, err = time.Parse(time.RFC3339, when)
		if err != nil {
			return when
		}
	}
	return t.Format("Jan 02 2006")
}
