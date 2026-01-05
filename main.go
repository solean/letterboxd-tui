package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const baseURL = "https://letterboxd.com"

type DiaryEntry struct {
	Date    string
	Title   string
	FilmURL string
	Rating  string
	Rewatch bool
	Review  bool
}

type Profile struct {
	Stats     []ProfileStat
	Favorites []FavoriteFilm
	Recent    []string
}

type ProfileStat struct {
	Label string
	Value string
	URL   string
}

type FavoriteFilm struct {
	Title   string
	FilmURL string
	Year    string
}

type ActivityItem struct {
	Summary string
	When    string
	Title   string
	FilmURL string
	Rating  string
	Kind    string
}

type tab int

const (
	tabProfile tab = iota
	tabDiary
	tabFollowing
	tabActivity
)

const tabCount = 4

type listState struct {
	selected int
}

type model struct {
	username    string
	cookie      string
	client      *http.Client
	width       int
	height      int
	activeTab   tab
	profile     Profile
	diary       []DiaryEntry
	activity    []ActivityItem
	following   []ActivityItem
	profileErr  error
	diaryErr    error
	activityErr error
	followErr   error
	loading     bool
	diaryList   listState
	actList     listState
	followList  listState
}

type diaryMsg struct {
	items []DiaryEntry
	err   error
}

type profileMsg struct {
	profile Profile
	err     error
}

type activityMsg struct {
	items []ActivityItem
	err   error
	tab   tab
}

type errMsg struct {
	err error
}

func main() {
	username := flag.String("user", "cschnabel", "Letterboxd username")
	flag.Parse()

	cookie, _ := loadCookie()
	client := &http.Client{Timeout: 12 * time.Second}

	m := model{
		username:  *username,
		cookie:    cookie,
		client:    client,
		activeTab: tabProfile,
		loading:   true,
	}

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

func (m model) Init() tea.Cmd {
	return tea.Batch(
		fetchProfileCmd(m.client, m.username, m.cookie),
		fetchDiaryCmd(m.client, m.username, m.cookie),
		fetchActivityCmd(m.client, m.username, m.cookie, tabActivity),
		fetchActivityCmd(m.client, m.username, m.cookie, tabFollowing),
	)
}

func fetchProfileCmd(client *http.Client, username, cookie string) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("%s/%s/", baseURL, username)
		doc, err := fetchDocument(client, url, cookie)
		if err != nil {
			return profileMsg{err: err}
		}
		profile, err := parseProfile(doc)
		return profileMsg{profile: profile, err: err}
	}
}

func fetchDiaryCmd(client *http.Client, username, cookie string) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("%s/%s/diary/", baseURL, username)
		doc, err := fetchDocument(client, url, cookie)
		if err != nil {
			return diaryMsg{err: err}
		}
		items, err := parseDiary(doc)
		return diaryMsg{items: items, err: err}
	}
}

func fetchActivityCmd(client *http.Client, username, cookie string, which tab) tea.Cmd {
	return func() tea.Msg {
		var url string
		switch which {
		case tabActivity:
			url = fmt.Sprintf("%s/ajax/activity-pagination/%s/", baseURL, username)
		case tabFollowing:
			url = followingActivityURL(username, cookie)
		default:
			return errMsg{err: fmt.Errorf("unknown activity tab")}
		}
		doc, err := fetchDocument(client, url, cookie)
		if err != nil {
			return activityMsg{tab: which, err: err}
		}
		items, err := parseActivity(doc)
		return activityMsg{tab: which, items: items, err: err}
	}
}

func fetchDocument(client *http.Client, url, cookie string) (*goquery.Document, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "letterboxd-tui/0.1")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}
	return goquery.NewDocumentFromReader(resp.Body)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab", "right":
			m.activeTab = (m.activeTab + 1) % tabCount
		case "left", "shift+tab":
			m.activeTab = (m.activeTab + tabCount - 1) % tabCount
		case "j", "down":
			m.moveSelection(1)
		case "k", "up":
			m.moveSelection(-1)
		case "r":
			m.loading = true
			return m, tea.Batch(
				fetchProfileCmd(m.client, m.username, m.cookie),
				fetchDiaryCmd(m.client, m.username, m.cookie),
				fetchActivityCmd(m.client, m.username, m.cookie, tabActivity),
				fetchActivityCmd(m.client, m.username, m.cookie, tabFollowing),
			)
		}
	case profileMsg:
		m.profile = msg.profile
		m.profileErr = msg.err
		m.loading = false
	case diaryMsg:
		m.diary = msg.items
		m.diaryErr = msg.err
		m.loading = false
	case activityMsg:
		if msg.tab == tabActivity {
			m.activity = msg.items
			m.activityErr = msg.err
		} else {
			m.following = msg.items
			m.followErr = msg.err
		}
		m.loading = false
	case errMsg:
		m.diaryErr = msg.err
		m.loading = false
	}
	return m, nil
}

func (m *model) moveSelection(delta int) {
	switch m.activeTab {
	case tabDiary:
		if len(m.diary) == 0 {
			return
		}
		m.diaryList.selected = clamp(m.diaryList.selected+delta, 0, len(m.diary)-1)
	case tabActivity:
		if len(m.activity) == 0 {
			return
		}
		m.actList.selected = clamp(m.actList.selected+delta, 0, len(m.activity)-1)
	case tabFollowing:
		if len(m.following) == 0 {
			return
		}
		m.followList.selected = clamp(m.followList.selected+delta, 0, len(m.following)-1)
	}
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

func (m model) View() string {
	theme := newTheme()
	header := theme.header.Render("Letterboxd TUI") + " " + theme.subtle.Render("@"+m.username)
	tabLine := renderTabs(m, theme)

	var body string
	switch m.activeTab {
	case tabProfile:
		body = renderProfile(m, theme)
	case tabDiary:
		body = renderDiary(m, theme)
	case tabActivity:
		body = renderActivity(m.activity, m.activityErr, m.actList.selected, theme)
	case tabFollowing:
		body = renderActivity(m.following, m.followErr, m.followList.selected, theme)
	}

	footer := theme.subtle.Render("tab/shift+tab switch • j/k move • r refresh • q quit")
	return lipgloss.JoinVertical(lipgloss.Left, header, tabLine, body, footer)
}

type themeStyles struct {
	header    lipgloss.Style
	subtle    lipgloss.Style
	tab       lipgloss.Style
	tabActive lipgloss.Style
	item      lipgloss.Style
	itemSel   lipgloss.Style
	badge     lipgloss.Style
	dim       lipgloss.Style
}

func newTheme() themeStyles {
	return themeStyles{
		header: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F2C14E")),
		subtle: lipgloss.NewStyle().Foreground(lipgloss.Color("#7A8C93")),
		tab:    lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#B0BEC5")),
		tabActive: lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("#0B0F0F")).
			Background(lipgloss.Color("#F2C14E")).
			Bold(true),
		item:    lipgloss.NewStyle().Padding(0, 1),
		itemSel: lipgloss.NewStyle().Padding(0, 1).Background(lipgloss.Color("#1C2B2D")).Foreground(lipgloss.Color("#E7E6E1")),
		badge:   lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#C9D1D5")).Background(lipgloss.Color("#2C3E42")),
		dim:     lipgloss.NewStyle().Foreground(lipgloss.Color("#55676E")),
	}
}

func renderTabs(m model, theme themeStyles) string {
	tabs := []string{"Profile", "Diary", "Friends", "My Activity"}
	var out []string
	for i, label := range tabs {
		if tab(i) == m.activeTab {
			out = append(out, theme.tabActive.Render(label))
		} else {
			out = append(out, theme.tab.Render(label))
		}
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, out...)
}

func renderProfile(m model, theme themeStyles) string {
	if m.profileErr != nil {
		return theme.dim.Render("Error: " + m.profileErr.Error())
	}
	if m.loading && len(m.profile.Stats) == 0 && len(m.profile.Favorites) == 0 {
		return theme.dim.Render("Loading profile…")
	}
	var rows []string
	if len(m.profile.Stats) > 0 {
		rows = append(rows, theme.subtle.Render("Stats"))
		for _, stat := range m.profile.Stats {
			line := fmt.Sprintf("%s %s", theme.badge.Render(stat.Value), stat.Label)
			rows = append(rows, theme.item.Render(line))
		}
	}
	if len(m.profile.Favorites) > 0 {
		rows = append(rows, "", theme.subtle.Render("Top 4 Films"))
		for i, fav := range m.profile.Favorites {
			prefix := fmt.Sprintf("%d.", i+1)
			title := fav.Title
			if fav.Year != "" {
				title = fmt.Sprintf("%s (%s)", fav.Title, fav.Year)
			}
			rows = append(rows, theme.item.Render(fmt.Sprintf("%s %s", prefix, title)))
		}
	}
	if len(m.profile.Recent) > 0 {
		rows = append(rows, "", theme.subtle.Render("Recently Watched"))
		for _, line := range m.profile.Recent {
			rows = append(rows, theme.item.Render(line))
		}
	}
	if len(rows) == 0 {
		return theme.dim.Render("No profile data found.")
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func renderDiary(m model, theme themeStyles) string {
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

func renderActivity(items []ActivityItem, err error, selected int, theme themeStyles) string {
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
		summary := compactSpaces(item.Summary)
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
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

func parseDiary(doc *goquery.Document) ([]DiaryEntry, error) {
	var entries []DiaryEntry
	currentMonth := ""
	currentYear := ""

	doc.Find("tr.diary-entry-row").Each(func(_ int, row *goquery.Selection) {
		month := strings.TrimSpace(row.Find(".col-monthdate .month").First().Text())
		year := strings.TrimSpace(row.Find(".col-monthdate .year").First().Text())
		day := strings.TrimSpace(row.Find(".col-daydate .daydate").First().Text())
		if month != "" {
			currentMonth = month
		}
		if year != "" {
			currentYear = year
		}

		titleSel := row.Find("h2.name a").First()
		title := strings.TrimSpace(titleSel.Text())
		filmURL, _ := titleSel.Attr("href")
		if filmURL != "" && strings.HasPrefix(filmURL, "/") {
			filmURL = baseURL + filmURL
		}
		rating := strings.TrimSpace(row.Find(".col-rating .rating").First().Text())
		rewatch := strings.Contains(row.Find(".js-td-rewatch").AttrOr("class", ""), "icon-status-on")
		review := row.Find(".js-td-review a").Length() > 0

		date := ""
		if day != "" && currentMonth != "" && currentYear != "" {
			date = fmt.Sprintf("%s %s %s", currentMonth, day, currentYear)
		}
		if title != "" {
			entries = append(entries, DiaryEntry{
				Date:    date,
				Title:   title,
				FilmURL: filmURL,
				Rating:  rating,
				Rewatch: rewatch,
				Review:  review,
			})
		}
	})
	return entries, nil
}

func parseActivity(doc *goquery.Document) ([]ActivityItem, error) {
	var items []ActivityItem
	doc.Find("section.activity-row").Each(func(_ int, row *goquery.Selection) {
		kind := strings.TrimSpace(row.AttrOr("class", ""))
		summary := strings.TrimSpace(row.Find(".activity-summary").First().Text())
		when := strings.TrimSpace(row.Find("time.time").First().AttrOr("datetime", ""))
		titleSel := row.Find("h2.name a").First()
		title := strings.TrimSpace(titleSel.Text())
		filmURL, _ := titleSel.Attr("href")
		if filmURL != "" && strings.HasPrefix(filmURL, "/") {
			filmURL = baseURL + filmURL
		}
		rating := strings.TrimSpace(row.Find(".rating").First().Text())
		items = append(items, ActivityItem{
			Summary: summary,
			When:    when,
			Title:   title,
			FilmURL: filmURL,
			Rating:  rating,
			Kind:    kind,
		})
	})
	return items, nil
}

func parseProfile(doc *goquery.Document) (Profile, error) {
	var profile Profile
	doc.Find(".profile-stats .profile-statistic").Each(func(_ int, stat *goquery.Selection) {
		value := strings.TrimSpace(stat.Find(".value").First().Text())
		label := strings.TrimSpace(stat.Find(".definition").First().Text())
		url, _ := stat.Find("a").First().Attr("href")
		if url != "" && strings.HasPrefix(url, "/") {
			url = baseURL + url
		}
		if value != "" && label != "" {
			profile.Stats = append(profile.Stats, ProfileStat{
				Label: label,
				Value: value,
				URL:   url,
			})
		}
	})

	doc.Find("#favourites .posteritem .react-component").Each(func(_ int, fav *goquery.Selection) {
		title := strings.TrimSpace(fav.AttrOr("data-item-name", ""))
		filmURL := strings.TrimSpace(fav.AttrOr("data-item-link", ""))
		if filmURL != "" && strings.HasPrefix(filmURL, "/") {
			filmURL = baseURL + filmURL
		}
		year := ""
		if open := strings.LastIndex(title, "("); open != -1 {
			if close := strings.LastIndex(title, ")"); close > open {
				year = strings.TrimSpace(title[open+1 : close])
				title = strings.TrimSpace(title[:open])
			}
		}
		if title != "" {
			profile.Favorites = append(profile.Favorites, FavoriteFilm{
				Title:   title,
				FilmURL: filmURL,
				Year:    year,
			})
		}
	})

	doc.Find("section.timeline .activity-summary").Each(func(_ int, summary *goquery.Selection) {
		line := compactSpaces(summary.Text())
		if line == "" {
			return
		}
		if !strings.Contains(strings.ToLower(line), "watched") {
			return
		}
		profile.Recent = append(profile.Recent, line)
	})
	return profile, nil
}

func followingActivityURL(username, cookie string) string {
	csrf := cookieValue(cookie, "com.xk72.webparts.csrf")
	query := "diaryEntries=true&reviews=true&lists=true&stories=true&reviewComments=true&listComments=true&storyComments=true&watchlistAdditions=true&reviewLikes=true&listLikes=true&storyLikes=true&follows=true&yourActivity=true&incomingActivity=true"
	if csrf != "" {
		query = query + "&__csrf=" + csrf
	}
	return fmt.Sprintf("%s/ajax/activity-pagination/%s/following/?%s", baseURL, username, query)
}

func cookieValue(cookie, key string) string {
	parts := strings.Split(cookie, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.HasPrefix(part, key+"=") {
			return strings.TrimPrefix(part, key+"=")
		}
	}
	return ""
}
