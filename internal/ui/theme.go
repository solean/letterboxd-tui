package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type themeStyles struct {
	header    lipgloss.Style
	subtle    lipgloss.Style
	tab       lipgloss.Style
	tabActive lipgloss.Style
	item      lipgloss.Style
	itemSel   lipgloss.Style
	badge     lipgloss.Style
	dim       lipgloss.Style
	user      lipgloss.Style
	movie     lipgloss.Style
	rateHigh  lipgloss.Style
	rateMid   lipgloss.Style
	rateLow   lipgloss.Style
}

func newTheme() themeStyles {
	return themeStyles{
		header: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00E054")),
		subtle: lipgloss.NewStyle().Foreground(lipgloss.Color("#9BB0B8")),
		tab:    lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#C9D1D5")),
		tabActive: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#14181C")).
			Background(lipgloss.Color("#00E054")).
			Bold(true),
		item:     lipgloss.NewStyle().Padding(0, 1),
		itemSel:  lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#1F2A33")).Foreground(lipgloss.Color("#E6F0F2")),
		badge:    lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#E6F0F2")).Background(lipgloss.Color("#2B3B45")),
		dim:      lipgloss.NewStyle().Foreground(lipgloss.Color("#7F8D96")),
		user:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C3A")).Bold(true),
		movie:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true),
		rateHigh: lipgloss.NewStyle().Foreground(lipgloss.Color("#00E054")).Bold(true),
		rateMid:  lipgloss.NewStyle().Foreground(lipgloss.Color("#F2C94C")).Bold(true),
		rateLow:  lipgloss.NewStyle().Foreground(lipgloss.Color("#E25555")).Bold(true),
	}
}

func styleRating(rating string, theme themeStyles) string {
	value := starsToValue(rating)
	switch {
	case value >= 5.0:
		return glowStars(rating)
	case value >= 4.0:
		return theme.rateHigh.Render(rating)
	case value >= 2.5:
		return theme.rateMid.Render(rating)
	case value > 0:
		return theme.rateLow.Render(rating)
	default:
		return rating
	}
}

func starsToValue(rating string) float64 {
	rating = strings.TrimSpace(rating)
	if rating == "" {
		return 0
	}
	var value float64
	for _, r := range rating {
		switch r {
		case '★':
			value += 1.0
		case '½':
			value += 0.5
		}
	}
	return value
}

func glowStars(rating string) string {
	gradient := []string{
		"#6BFF6A",
		"#7BFF5A",
		"#8CFF4A",
		"#9EFF3A",
		"#B0FF2A",
	}
	var out strings.Builder
	colorIndex := 0
	for _, r := range rating {
		switch r {
		case '★', '½':
			color := gradient[min(colorIndex, len(gradient)-1)]
			out.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Bold(true).Render(string(r)))
			colorIndex++
		default:
			out.WriteRune(r)
		}
	}
	return out.String()
}
