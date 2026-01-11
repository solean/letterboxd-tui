package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"letterboxd-tui/internal/letterboxd"
)

func (m Model) View() string {
	theme := newTheme()
	header := theme.header.Render("Letterboxd TUI") + " " + theme.subtle.Render("@"+m.username)
	tabLine := renderTabs(m, theme)

	var body string
	switch m.activeTab {
	case tabProfile:
		body = renderProfile(m, theme)
	case tabDiary:
		body = renderDiary(m, theme)
	case tabWatchlist:
		body = renderWatchlist(m, theme)
	case tabFilm:
		body = renderFilm(m, theme)
	case tabActivity:
		body = renderActivity(m.activity, m.activityErr, m.actList.selected, theme)
	case tabFollowing:
		body = renderActivity(m.following, m.followErr, m.followList.selected, theme)
	}

	footer := theme.subtle.Render(renderLegend(m))
	vp := m.viewport
	vp.SetContent(body)
	base := lipgloss.JoinVertical(lipgloss.Left, header, tabLine, vp.View(), footer)
	if m.activeTab == tabFilm {
		return renderFilmModal(base, m, theme)
	}
	if m.profileModal {
		return renderProfileModal(base, m, theme)
	}
	return base
}

func renderTabs(m Model, theme themeStyles) string {
	tabs := []string{"Profile", "Diary", "Friends", "My Activity", "Watchlist"}
	var out []string
	for i, label := range tabs {
		if visibleTabByIndex(i) == m.activeTab {
			out = append(out, theme.tabActive.Render(label))
		} else {
			out = append(out, theme.tab.Render(label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, out...)
}

func visibleTabs() []tab {
	return []tab{tabProfile, tabDiary, tabFollowing, tabActivity, tabWatchlist}
}

func visibleTabByIndex(i int) tab {
	tabs := visibleTabs()
	if i < 0 || i >= len(tabs) {
		return tabProfile
	}
	return tabs[i]
}

func nextTab(current tab) tab {
	if current == tabFilm {
		return tabProfile
	}
	tabs := visibleTabs()
	for i, t := range tabs {
		if t == current {
			return tabs[(i+1)%len(tabs)]
		}
	}
	return tabProfile
}

func prevTab(current tab) tab {
	if current == tabFilm {
		return tabProfile
	}
	tabs := visibleTabs()
	for i, t := range tabs {
		if t == current {
			if i == 0 {
				return tabs[len(tabs)-1]
			}
			return tabs[i-1]
		}
	}
	return tabProfile
}

func renderLegend(m Model) string {
	if m.activeTab == tabFilm {
		return "j/k scroll • pgup/pgdn page • o open film in browser • tab/shift+tab back • esc/q close film"
	}
	if m.profileModal {
		return "j/k scroll • pgup/pgdn page • o open in browser • b/q/esc close profile"
	}

	var parts []string
	parts = append(parts, "tab/shift+tab switch")
	if m.activeTab == tabProfile {
		parts = append(parts, "j/k scroll", "pgup/pgdn scroll")
		if len(m.profileStack) > 0 {
			parts = append(parts, "b back")
		}
		parts = append(parts, "o open profile in browser")
	} else {
		parts = append(parts, "j/k move", "pgup/pgdn page")
		switch m.activeTab {
		case tabDiary, tabWatchlist, tabActivity:
			parts = append(parts, "enter view film")
		case tabFollowing:
			parts = append(parts, "enter view profile")
		}
	}
	parts = append(parts, "r refresh", "q quit")
	return strings.Join(parts, " • ")
}

func renderProfile(m Model, theme themeStyles) string {
	return renderProfileContent(m.profile, m.profileErr, m.loading, m.profileUser, m.profileStack, theme)
}

func renderDiary(m Model, theme themeStyles) string {
	if m.diaryErr != nil {
		return theme.dim.Render("Error: " + m.diaryErr.Error())
	}
	if m.loading && len(m.diary) == 0 {
		return theme.dim.Render("Loading diary…")
	}
	if len(m.diary) == 0 {
		return theme.dim.Render("No diary entries found.")
	}
	var rows []string
	width := max(40, m.width-2)
	for i, entry := range m.diary {
		date := theme.badge.Render(entry.Date)
		rating := entry.Rating
		if rating == "" {
			rating = "—"
		} else {
			rating = styleRating(rating, theme)
		}
		flags := ""
		if entry.Rewatch {
			flags += " ↺"
		}
		if entry.Review {
			flags += " ✎"
		}
		line := fmt.Sprintf("%s %s %s%s", date, entry.Title, rating, flags)
		line = truncate(line, width)
		style := theme.item
		if i == m.diaryList.selected {
			style = theme.itemSel
		}
		rows = append(rows, style.Render(line))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderWatchlist(m Model, theme themeStyles) string {
	if m.watchErr != nil {
		return theme.dim.Render("Error: " + m.watchErr.Error())
	}
	if m.loading && len(m.watchlist) == 0 {
		return theme.dim.Render("Loading watchlist…")
	}
	if len(m.watchlist) == 0 {
		return theme.dim.Render("No watchlist items found.")
	}
	var rows []string
	width := max(40, m.width-2)
	for i, item := range m.watchlist {
		title := item.Title
		if item.Year != "" {
			title = fmt.Sprintf("%s (%s)", item.Title, item.Year)
		}
		line := truncate(title, width)
		style := theme.item
		if i == m.watchList.selected {
			style = theme.itemSel
		}
		rows = append(rows, style.Render(line))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderProfileContent(profile letterboxd.Profile, err error, loading bool, profileUser string, stack []string, theme themeStyles) string {
	if err != nil {
		return theme.dim.Render("Error: " + err.Error())
	}
	if loading && len(profile.Stats) == 0 && len(profile.Favorites) == 0 {
		return theme.dim.Render("Loading profile…")
	}
	var rows []string
	rows = append(rows, theme.subtle.Render(renderBreadcrumbs(stack, profileUser, theme.user)))
	if len(profile.Stats) > 0 {
		rows = append(rows, "", theme.subtle.Render("Stats"))
		for _, stat := range profile.Stats {
			line := fmt.Sprintf("%s %s", theme.badge.Render(stat.Value), stat.Label)
			rows = append(rows, theme.item.Render(line))
		}
	}
	if len(profile.Favorites) > 0 {
		rows = append(rows, "", theme.subtle.Render("Top 4 Films"))
		for i, fav := range profile.Favorites {
			prefix := fmt.Sprintf("%d.", i+1)
			title := fav.Title
			if fav.Year != "" {
				title = fmt.Sprintf("%s (%s)", fav.Title, fav.Year)
			}
			rows = append(rows, theme.item.Render(fmt.Sprintf("%s %s", prefix, title)))
		}
	}
	if len(profile.Recent) > 0 {
		rows = append(rows, "", theme.subtle.Render("Recently Watched"))
		for _, line := range profile.Recent {
			line = strings.Replace(line, profileUser, theme.user.Render(profileUser), 1)
			rows = append(rows, theme.item.Render(line))
		}
	}
	if len(rows) == 0 {
		return theme.dim.Render("No profile data found.")
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderFilm(m Model, theme themeStyles) string {
	if m.filmErr != nil {
		return theme.dim.Render("Error: " + m.filmErr.Error())
	}
	if m.loading && m.film.Title == "" {
		return theme.dim.Render("Loading film…")
	}
	if m.film.Title == "" {
		return theme.dim.Render("No film details found.")
	}
	var rows []string
	title := m.film.Title
	if m.film.Year != "" {
		title = fmt.Sprintf("%s (%s)", title, m.film.Year)
	}
	rows = append(rows, theme.header.Render(title))
	meta := []string{}
	if m.film.Director != "" {
		meta = append(meta, "Dir. "+m.film.Director)
	}
	if m.film.Runtime != "" {
		meta = append(meta, m.film.Runtime)
	}
	if m.film.AvgRating != "" {
		meta = append(meta, "Avg "+m.film.AvgRating)
	}
	if len(meta) > 0 {
		rows = append(rows, theme.subtle.Render(strings.Join(meta, " • ")))
	}
	if m.film.UserStatus != "" || m.film.UserRating != "" {
		userLine := "You: "
		if m.film.UserStatus != "" {
			userLine += m.film.UserStatus
		}
		if m.film.UserRating != "" {
			if m.film.UserStatus != "" {
				userLine += " "
			}
			userLine += styleRating(m.film.UserRating, theme)
		}
		rows = append(rows, theme.subtle.Render(userLine))
	}
	if len(m.film.Cast) > 0 {
		rows = append(rows, "", theme.subtle.Render("Top Billed Cast"))
		rows = append(rows, theme.item.Render(strings.Join(m.film.Cast, ", ")))
	}
	if m.film.Description != "" {
		rows = append(rows, "", wrapText(m.film.Description, max(40, m.width-4)))
	}
	if m.film.URL != "" {
		rows = append(rows, "", theme.dim.Render(m.film.URL))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderFilmModal(base string, m Model, theme themeStyles) string {
	overlay := renderFilm(m, theme)
	if overlay == "" {
		return base
	}
	width, height := modalDimensions(m.width, m.height)
	innerWidth := width - 4
	innerHeight := height - 2
	legend := theme.subtle.Render(renderLegend(m))
	legendHeight := lipgloss.Height(legend)
	bodyHeight := max(0, innerHeight-legendHeight-1)

	body := lipgloss.Place(innerWidth, bodyHeight, lipgloss.Left, lipgloss.Top, overlay)
	content := lipgloss.JoinVertical(lipgloss.Left, body, "", legend)
	panel := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3A4A55")).
		Background(lipgloss.Color("#14181C")).
		Foreground(lipgloss.Color("#E6F0F2"))
	panelContent := lipgloss.Place(innerWidth, innerHeight, lipgloss.Left, lipgloss.Top, content)
	modal := panel.Render(panelContent)

	dim := lipgloss.NewStyle().
		Background(lipgloss.Color("#0E1114")).
		Foreground(lipgloss.Color("#5E6A72")).
		Render(base)
	return dim + "\n" + lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceBackground(lipgloss.Color("#0E1114")))
}

func renderProfileModal(base string, m Model, theme themeStyles) string {
	overlay := renderProfileContent(m.modalProfile, m.modalProfileErr, m.modalLoading, m.modalUser, nil, theme)
	if overlay == "" {
		return base
	}
	width, height := modalDimensions(m.width, m.height)
	innerWidth := width - 4
	innerHeight := height - 2
	legend := theme.subtle.Render(renderLegend(m))
	legendHeight := lipgloss.Height(legend)
	bodyHeight := max(0, innerHeight-legendHeight-1)

	body := lipgloss.Place(innerWidth, bodyHeight, lipgloss.Left, lipgloss.Top, overlay)
	content := lipgloss.JoinVertical(lipgloss.Left, body, "", legend)
	panel := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3A4A55")).
		Background(lipgloss.Color("#14181C")).
		Foreground(lipgloss.Color("#E6F0F2")).
		Padding(1, 2)
	panelContent := lipgloss.Place(innerWidth, innerHeight, lipgloss.Left, lipgloss.Top, content)
	modal := panel.Render(panelContent)

	dim := lipgloss.NewStyle().
		Background(lipgloss.Color("#0E1114")).
		Foreground(lipgloss.Color("#5E6A72")).
		Render(base)
	return dim + "\n" + lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceBackground(lipgloss.Color("#0E1114")))
}

func renderActivity(items []letterboxd.ActivityItem, err error, selected int, theme themeStyles) string {
	if err != nil {
		return theme.dim.Render("Error: " + err.Error())
	}
	if len(items) == 0 {
		return theme.dim.Render("No activity found.")
	}
	var rows []string
	for i, item := range items {
		when := formatWhen(item.When)
		if when == "" {
			when = "—"
		}
		summary := renderSummary(item, theme)
		if summary == "" {
			summary = item.Title
		}
		line := fmt.Sprintf("%s %s", theme.badge.Render(when), summary)
		style := theme.item
		if i == selected {
			style = theme.itemSel
		}
		rows = append(rows, style.Render(line))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderSummary(item letterboxd.ActivityItem, theme themeStyles) string {
	if len(item.Parts) == 0 {
		return compactSpaces(item.Summary)
	}
	var out strings.Builder
	for _, part := range item.Parts {
		text := compactSpaces(part.Text)
		if text == "" {
			continue
		}
		switch part.Kind {
		case "user":
			text = theme.user.Render(text)
		case "movie":
			text = theme.movie.Render(text)
		case "rating":
			text = styleRating(text, theme)
		}
		appendWithSpacing(&out, text)
	}
	return strings.TrimSpace(out.String())
}

func renderBreadcrumbs(stack []string, current string, userStyle lipgloss.Style) string {
	if current == "" {
		return "Profile"
	}
	parts := make([]string, 0, len(stack)+1)
	for _, name := range stack {
		if name != "" {
			parts = append(parts, userStyle.Render("@"+name))
		}
	}
	parts = append(parts, userStyle.Render("@"+current))
	return "Profile: " + strings.Join(parts, " > ")
}
