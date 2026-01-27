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
		body = renderActivityWithStatus(m.activity, m.activityErr, m.activityMoreErr, m.actList.selected, m.width, m.activityLoadingMore, m.activityDone, theme)
	case tabFollowing:
		body = renderActivityWithStatus(m.following, m.followErr, m.followMoreErr, m.followList.selected, m.width, m.followLoadingMore, m.followDone, theme)
	case tabSearch:
		body = renderSearch(m, theme)
	}

	footer := renderHelp(m, theme, m.width)
	vp := m.viewport
	vp.SetContent(body)
	base := lipgloss.JoinVertical(lipgloss.Left, header, tabLine, vp.View(), footer)
	if m.activeTab == tabFilm {
		base = renderFilmModal(base, m, theme)
	}
	if m.profileModal {
		base = renderProfileModal(base, m, theme)
	}
	if m.logModal {
		base = renderLogModal(base, m, theme)
	}
	return base
}

func renderTabs(m Model, theme themeStyles) string {
	tabs := []string{"Profile", "Diary", "Friends", "My Activity", "Watchlist", "Search"}
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
	return []tab{tabProfile, tabDiary, tabFollowing, tabActivity, tabWatchlist, tabSearch}
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
		rows = append(rows, renderSelectableLine(line, i == m.diaryList.selected, width, theme))
	}
	if status := renderListStatus(m.diaryLoadingMore, m.diaryMoreErr, m.diaryDone, theme); status != "" {
		rows = append(rows, status)
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
		rows = append(rows, renderSelectableLine(title, i == m.watchList.selected, width, theme))
	}
	if status := renderListStatus(m.watchLoadingMore, m.watchMoreErr, m.watchDone, theme); status != "" {
		rows = append(rows, status)
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

func renderSearch(m Model, theme themeStyles) string {
	var rows []string
	rows = append(rows, theme.subtle.Render("Search films"))

	line := "Query: " + m.searchInput.View()
	if m.searchFocusInput {
		rows = append(rows, theme.itemSel.Render(line))
	} else {
		rows = append(rows, theme.item.Render(line))
	}

	if m.searchLoading {
		rows = append(rows, theme.dim.Render("Searching…"))
	}
	if m.searchErr != nil {
		rows = append(rows, theme.dim.Render("Error: "+m.searchErr.Error()))
	}
	if !m.searchLoading && m.searchErr == nil && len(m.searchResults) == 0 {
		rows = append(rows, theme.dim.Render("No results yet."))
	}

	width := max(40, m.width-2)
	for i, r := range m.searchResults {
		title := r.Title
		if r.Year != "" {
			title = fmt.Sprintf("%s (%s)", r.Title, r.Year)
		}
		selected := i == m.searchList.selected && !m.searchFocusInput
		rows = append(rows, renderSelectableLine(title, selected, width, theme))
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
	wrapWidth := max(40, modalContentWidth(m.width, m.height)-2)
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
			rows = append(rows, theme.subtle.Render(userLine))
		}
	}
	if inWatchlist, ok := m.watchlistState(); ok {
		label := ""
		style := theme.subtle
		if inWatchlist {
			label = "On your watchlist"
			style = theme.rateHigh
		}
		rows = append(rows, style.Render(label))
	}
	if m.watchlistStatus != "" {
		rows = append(rows, renderWatchlistStatus(m.watchlistStatus, theme))
	}
	if len(m.film.Cast) > 0 {
		rows = append(rows, "", theme.subtle.Render("Top Billed Cast"))
		rows = append(rows, theme.item.Render(wrapText(strings.Join(m.film.Cast, ", "), wrapWidth)))
	}
	if m.film.Description != "" {
		rows = append(rows, "", wrapText(m.film.Description, wrapWidth))
	}
	if m.film.URL != "" {
		rows = append(rows, "", theme.dim.Render(truncate(m.film.URL, wrapWidth)))
	}
	if len(m.friendReviews) > 0 || m.friendReviewsErr != nil {
		rows = append(rows, "", renderReviewsWithStatus("Friends' reviews", m.friendReviews, m.friendReviewsErr, wrapWidth, m.friendReviewsLoadingMore, m.friendReviewsMoreErr, m.friendReviewsDone, theme))
	}
	if len(m.popReviews) > 0 || m.popReviewsErr != nil {
		rows = append(rows, "", renderReviewsWithStatus("Popular reviews", m.popReviews, m.popReviewsErr, wrapWidth, m.popReviewsLoadingMore, m.popReviewsMoreErr, m.popReviewsDone, theme))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderFilmModal(base string, m Model, theme themeStyles) string {
	if m.modalVP.View() == "" && renderFilm(m, theme) == "" {
		return base
	}
	width, height := modalDimensions(m.width, m.height)
	innerWidth := width - 4
	innerHeight := height - 2
	legend := renderHelp(m, theme, innerWidth)
	legendHeight := lipgloss.Height(legend)
	bodyHeight := max(1, innerHeight-legendHeight-1)

	vp := m.modalVP
	vp.Width = innerWidth
	vp.Height = bodyHeight
	body := vp.View()
	if body == "" {
		vp.SetContent(renderFilm(m, theme))
		body = vp.View()
	}
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
	if m.modalVP.View() == "" && renderProfileContent(m.modalProfile, m.modalProfileErr, m.modalLoading, m.modalUser, nil, theme) == "" {
		return base
	}
	width, height := modalDimensions(m.width, m.height)
	innerWidth := width - 4
	innerHeight := height - 2
	legend := renderHelp(m, theme, innerWidth)
	legendHeight := lipgloss.Height(legend)
	bodyHeight := max(1, innerHeight-legendHeight-1)

	vp := m.modalVP
	vp.Width = innerWidth
	vp.Height = bodyHeight
	body := vp.View()
	if body == "" {
		vp.SetContent(renderProfileContent(m.modalProfile, m.modalProfileErr, m.modalLoading, m.modalUser, nil, theme))
		body = vp.View()
	}
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

func renderLogModal(base string, m Model, theme themeStyles) string {
	form := renderLogForm(m, theme)
	width, height := modalDimensions(m.width, m.height)
	innerWidth := width - 4
	innerHeight := height - 2
	legend := renderHelp(m, theme, innerWidth)
	bodyHeight := max(0, innerHeight-lipgloss.Height(legend)-1)
	bodyHeight = max(1, bodyHeight)
	body := lipgloss.Place(innerWidth, bodyHeight, lipgloss.Left, lipgloss.Top, form)
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

func renderLogForm(m Model, theme themeStyles) string {
	var rows []string
	titleLine := m.film.Title
	if m.film.Year != "" {
		titleLine = fmt.Sprintf("%s (%s)", m.film.Title, m.film.Year)
	}
	rows = append(rows, theme.header.Render("Log diary entry"), theme.subtle.Render(titleLine))
	if status := renderLogStatus(m, theme); status != "" {
		rows = append(rows, status)
	}

	rows = append(rows, renderLogInput("Rating", m.logForm.rating.View(), m.logForm.focus == logFieldRating, theme))
	rows = append(rows, renderLogToggle("Rewatch", m.logForm.rewatch, m.logForm.focus == logFieldRewatch, theme))
	rows = append(rows, renderLogInput("Watched date", m.logForm.date.View(), m.logForm.focus == logFieldDate, theme))

	review := m.logForm.review.View()
	if m.logForm.focus == logFieldReview {
		review = theme.itemSel.Render(review)
	} else {
		review = theme.item.Render(review)
	}
	rows = append(rows, theme.subtle.Render("Review"), review)

	rows = append(rows, renderLogToggle("Contains spoilers", m.logForm.spoilers, m.logForm.focus == logFieldSpoilers, theme))
	rows = append(rows, renderLogToggle("Like", m.logForm.liked, m.logForm.focus == logFieldLiked, theme))
	rows = append(rows, renderLogInput("Tags", m.logForm.tags.View(), m.logForm.focus == logFieldTags, theme))
	rows = append(rows, renderLogInput("Privacy", m.logForm.privacyLabel(), m.logForm.focus == logFieldPrivacy, theme))
	rows = append(rows, renderLogToggle("Draft", m.logForm.draft, m.logForm.focus == logFieldDraft, theme))

	submitLabel := "Submit"
	if m.logForm.submitting {
		submitLabel = "Submitting…"
	}
	submit := theme.item.Render(submitLabel)
	if m.logForm.focus == logFieldSubmit {
		submit = theme.itemSel.Render(submitLabel)
	}
	rows = append(rows, submit)
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderLogStatus(m Model, theme themeStyles) string {
	if m.logForm.submitting {
		return theme.subtle.Render(m.logSpinner.View() + " Submitting…")
	}
	if m.logForm.status == "" {
		return ""
	}
	if strings.HasPrefix(m.logForm.status, "Error:") {
		return theme.rateLow.Render(m.logForm.status)
	}
	return theme.rateHigh.Render(m.logForm.status)
}

func renderWatchlistStatus(status string, theme themeStyles) string {
	if status == "" {
		return ""
	}
	if strings.HasPrefix(status, "Error:") {
		return theme.rateLow.Render(status)
	}
	if strings.HasPrefix(status, "Adding") || strings.HasPrefix(status, "Removing") {
		return theme.subtle.Render(status)
	}
	return theme.rateHigh.Render(status)
}

func renderLogInput(label, value string, focused bool, theme themeStyles) string {
	style := theme.item
	if focused {
		style = theme.itemSel
	}
	return style.Render(label + ": " + value)
}

func renderLogToggle(label string, on bool, focused bool, theme themeStyles) string {
	state := "no"
	if on {
		state = "yes"
	}
	style := theme.item
	if focused {
		style = theme.itemSel
	}
	return style.Render(label + ": " + state)
}

func renderSelectableLine(line string, selected bool, width int, theme themeStyles) string {
	prefix := "  "
	style := theme.item
	if selected {
		prefix = "> "
		style = theme.itemSel
	}
	line = prefix + line
	if width > 0 {
		line = truncate(line, width)
	}
	return style.Render(line)
}

func renderReviews(title string, reviews []letterboxd.Review, err error, width int, theme themeStyles) string {
	if err != nil {
		return theme.dim.Render("Error: " + err.Error())
	}
	if len(reviews) == 0 {
		return theme.dim.Render("No reviews found.")
	}
	var rows []string
	rows = append(rows, theme.subtle.Render(title))
	for _, r := range reviews {
		line := theme.user.Render(r.Author)
		if r.Rating != "" {
			line += " " + styleRating(r.Rating, theme)
		}
		rows = append(rows, truncate(line, width))
		if r.Text != "" {
			rows = append(rows, wrapText(r.Text, width))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderReviewsWithStatus(title string, reviews []letterboxd.Review, err error, width int, loadingMore bool, moreErr error, done bool, theme themeStyles) string {
	body := renderReviews(title, reviews, err, width, theme)
	if err != nil || len(reviews) == 0 {
		return body
	}
	status := renderListStatus(loadingMore, moreErr, done, theme)
	if status == "" {
		return body
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, status)
}

func renderActivity(items []letterboxd.ActivityItem, err error, selected int, width int, theme themeStyles) string {
	if err != nil {
		return theme.dim.Render("Error: " + err.Error())
	}
	if len(items) == 0 {
		return theme.dim.Render("No activity found.")
	}
	var rows []string
	width = max(40, width-2)
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
		rows = append(rows, renderSelectableLine(line, i == selected, width, theme))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderActivityWithStatus(items []letterboxd.ActivityItem, err, moreErr error, selected int, width int, loadingMore bool, done bool, theme themeStyles) string {
	body := renderActivity(items, err, selected, width, theme)
	if err != nil || len(items) == 0 {
		return body
	}
	status := renderListStatus(loadingMore, moreErr, done, theme)
	if status == "" {
		return body
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, status)
}

func renderListStatus(loading bool, moreErr error, done bool, theme themeStyles) string {
	if loading {
		return theme.dim.Render("Loading more…")
	}
	if moreErr != nil {
		return theme.rateLow.Render("Error loading more: " + moreErr.Error())
	}
	if done {
		return theme.dim.Render("End of list.")
	}
	return ""
}

func renderSummary(item letterboxd.ActivityItem, theme themeStyles) string {
	if len(item.Parts) == 0 {
		if s := compactSpaces(item.Summary); s != "" {
			return s
		}
		var parts []string
		if item.Actor != "" {
			parts = append(parts, theme.user.Render(item.Actor))
		}
		if verb := describeKind(item.Kind); verb != "" {
			parts = append(parts, verb)
		}
		if item.Title != "" {
			parts = append(parts, theme.movie.Render(item.Title))
		}
		if item.Rating != "" {
			parts = append(parts, styleRating(item.Rating, theme))
		}
		return strings.Join(parts, " ")
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

func describeKind(kind string) string {
	kind = strings.ToLower(kind)
	switch {
	case strings.Contains(kind, "watchlist"):
		return "added to watchlist"
	case strings.Contains(kind, "like"):
		return "liked"
	case strings.Contains(kind, "review"):
		return "reviewed"
	case strings.Contains(kind, "diary") || strings.Contains(kind, "watch"):
		return "watched"
	case strings.Contains(kind, "list"):
		return "listed"
	default:
		return ""
	}
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
