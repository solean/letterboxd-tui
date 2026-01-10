package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/net/html"
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

type WatchlistItem struct {
	Title   string
	FilmURL string
	Year    string
}

type Film struct {
	Title       string
	Year        string
	Description string
	Director    string
	AvgRating   string
	Runtime     string
	URL         string
	Cast        []string
	UserRating  string
	UserStatus  string
}

type ActivityItem struct {
	Summary  string
	When     string
	Title    string
	FilmURL  string
	Rating   string
	Kind     string
	Actor    string
	ActorURL string
	Parts    []SummaryPart
}

type SummaryPart struct {
	Text string
	Kind string
}

type tab int

const (
	tabProfile tab = iota
	tabDiary
	tabWatchlist
	tabFilm
	tabFollowing
	tabActivity
)

const tabCount = 5

type listState struct {
	selected int
}

type model struct {
	username     string
	profileUser  string
	profileStack []string
	cookie       string
	client       *http.Client
	width        int
	height       int
	activeTab    tab
	lastTab      tab
	profile      Profile
	diary        []DiaryEntry
	watchlist    []WatchlistItem
	activity     []ActivityItem
	following    []ActivityItem
	film         Film
	profileErr   error
	diaryErr     error
	watchErr     error
	activityErr  error
	followErr    error
	filmErr      error
	loading      bool
	diaryList    listState
	watchList    listState
	actList      listState
	followList   listState
	viewport     viewport.Model
	modalVP      viewport.Model
	filmReturn   tab
	profileModal bool
}

type diaryMsg struct {
	items []DiaryEntry
	err   error
}

type watchlistMsg struct {
	items []WatchlistItem
	err   error
}

type filmMsg struct {
	film Film
	err  error
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

type openMsg struct {
	err error
}

func main() {
	username := flag.String("user", "cschnabel", "Letterboxd username")
	flag.Parse()

	cookie, _ := loadCookie()
	client := &http.Client{Timeout: 12 * time.Second}

	m := model{
		username:    *username,
		profileUser: *username,
		cookie:      cookie,
		client:      client,
		activeTab:   tabProfile,
		lastTab:     tabProfile,
		loading:     true,
		viewport:    viewport.New(0, 0),
		modalVP:     viewport.New(0, 0),
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
		fetchProfileCmd(m.client, m.profileUser, m.cookie),
		fetchDiaryCmd(m.client, m.username, m.cookie),
		fetchWatchlistCmd(m.client, m.username, m.cookie),
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

func fetchWatchlistCmd(client *http.Client, username, cookie string) tea.Cmd {
	return func() tea.Msg {
		url := fmt.Sprintf("%s/%s/watchlist/", baseURL, username)
		doc, err := fetchDocument(client, url, cookie)
		if err != nil {
			return watchlistMsg{err: err}
		}
		items, err := parseWatchlist(doc)
		return watchlistMsg{items: items, err: err}
	}
}

func fetchFilmCmd(client *http.Client, filmURL, username, cookie string) tea.Cmd {
	return func() tea.Msg {
		doc, err := fetchDocument(client, filmURL, cookie)
		if err != nil {
			return filmMsg{err: err}
		}
		film, err := parseFilm(doc, filmURL)
		if err == nil && username != "" {
			userURL := userFilmURL(username, filmURL)
			if userURL != "" {
				userDoc, status, err := fetchDocumentAllowStatus(client, userURL, cookie)
				if err == nil && status == http.StatusOK {
					film.UserRating, film.UserStatus = parseUserFilm(userDoc)
				}
			}
		}
		return filmMsg{film: film, err: err}
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

func fetchDocumentAllowStatus(client *http.Client, url, cookie string) (*goquery.Document, int, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("User-Agent", "letterboxd-tui/0.1")
	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp.StatusCode, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	return doc, resp.StatusCode, err
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		bodyHeight := max(1, m.height-3)
		m.viewport.Width = m.width
		m.viewport.Height = bodyHeight
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.modalOpen() {
				if m.activeTab == tabFilm {
					m.activeTab = m.filmReturn
					m.resetTabPosition()
				} else if m.profileModal {
					m.profileModal = false
				}
				return m, nil
			}
			return m, tea.Quit
		case "tab", "right":
			if m.activeTab == tabFilm {
				m.activeTab = m.filmReturn
			} else {
				m.activeTab = nextTab(m.activeTab)
			}
			m.resetTabPosition()
		case "left", "shift+tab":
			if m.activeTab == tabFilm {
				m.activeTab = m.filmReturn
			} else {
				m.activeTab = prevTab(m.activeTab)
			}
			m.resetTabPosition()
		case "j", "down":
			if m.modalOpen() {
				m.modalVP.LineDown(1)
			} else if m.activeTab == tabProfile {
				m.viewport.LineDown(1)
			} else {
				m.moveSelection(1)
				m.syncViewportToSelection()
			}
		case "k", "up":
			if m.modalOpen() {
				m.modalVP.LineUp(1)
			} else if m.activeTab == tabProfile {
				m.viewport.LineUp(1)
			} else {
				m.moveSelection(-1)
				m.syncViewportToSelection()
			}
		case "pgdown":
			if m.modalOpen() {
				m.modalVP.ViewDown()
			} else if m.activeTab == tabProfile {
				m.viewport.ViewDown()
			} else {
				m.pageSelection(1)
				m.syncViewportToSelection()
			}
		case "pgup":
			if m.modalOpen() {
				m.modalVP.ViewUp()
			} else if m.activeTab == tabProfile {
				m.viewport.ViewUp()
			} else {
				m.pageSelection(-1)
				m.syncViewportToSelection()
			}
		case "r":
			m.loading = true
			return m, tea.Batch(
				fetchProfileCmd(m.client, m.profileUser, m.cookie),
				fetchDiaryCmd(m.client, m.username, m.cookie),
				fetchWatchlistCmd(m.client, m.username, m.cookie),
				fetchActivityCmd(m.client, m.username, m.cookie, tabActivity),
				fetchActivityCmd(m.client, m.username, m.cookie, tabFollowing),
			)
		case "enter":
			if m.activeTab == tabFollowing {
				m = m.openSelectedProfile()
				return m, fetchProfileCmd(m.client, m.profileUser, m.cookie)
			} else if m.activeTab == tabDiary || m.activeTab == tabWatchlist || m.activeTab == tabActivity {
				m = m.openSelectedFilm()
				if m.activeTab == tabFilm {
					return m, fetchFilmCmd(m.client, m.film.URL, m.username, m.cookie)
				}
			}
		case "b":
			if m.profileModal {
				m.profileModal = false
			} else if m.activeTab == tabProfile {
				m = m.goBackProfile()
				if m.activeTab == tabProfile {
					return m, fetchProfileCmd(m.client, m.profileUser, m.cookie)
				}
			}
		case "o":
			if m.profileModal || m.activeTab == tabProfile {
				return m, openBrowserCmd(profileURL(m.profileUser))
			}
		case "esc":
			if m.activeTab == tabFilm {
				m.activeTab = m.filmReturn
				m.resetTabPosition()
			} else if m.profileModal {
				m.profileModal = false
			}
		}
	case profileMsg:
		m.profile = msg.profile
		m.profileErr = msg.err
		m.loading = false
	case diaryMsg:
		m.diary = msg.items
		m.diaryErr = msg.err
		m.loading = false
	case watchlistMsg:
		m.watchlist = msg.items
		m.watchErr = msg.err
		m.loading = false
	case filmMsg:
		m.film = msg.film
		m.filmErr = msg.err
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
	case openMsg:
		if msg.err != nil {
			m.profileErr = msg.err
		}
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
	case tabWatchlist:
		if len(m.watchlist) == 0 {
			return
		}
		m.watchList.selected = clamp(m.watchList.selected+delta, 0, len(m.watchlist)-1)
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

func (m *model) resetTabPosition() {
	if m.activeTab == m.lastTab {
		return
	}
	m.viewport.YOffset = 0
	switch m.activeTab {
	case tabDiary:
		m.diaryList.selected = 0
	case tabWatchlist:
		m.watchList.selected = 0
	case tabFilm:
		m.filmErr = nil
	case tabActivity:
		m.actList.selected = 0
	case tabFollowing:
		m.followList.selected = 0
	}
	m.lastTab = m.activeTab
}

func (m *model) pageSelection(dir int) {
	step := max(1, m.viewport.Height-1)
	switch m.activeTab {
	case tabDiary:
		if len(m.diary) == 0 {
			return
		}
		m.diaryList.selected = clamp(m.diaryList.selected+dir*step, 0, len(m.diary)-1)
	case tabWatchlist:
		if len(m.watchlist) == 0 {
			return
		}
		m.watchList.selected = clamp(m.watchList.selected+dir*step, 0, len(m.watchlist)-1)
	case tabActivity:
		if len(m.activity) == 0 {
			return
		}
		m.actList.selected = clamp(m.actList.selected+dir*step, 0, len(m.activity)-1)
	case tabFollowing:
		if len(m.following) == 0 {
			return
		}
		m.followList.selected = clamp(m.followList.selected+dir*step, 0, len(m.following)-1)
	}
}

func (m *model) syncViewportToSelection() {
	var total, selected int
	switch m.activeTab {
	case tabDiary:
		total = len(m.diary)
		selected = m.diaryList.selected
	case tabWatchlist:
		total = len(m.watchlist)
		selected = m.watchList.selected
	case tabActivity:
		total = len(m.activity)
		selected = m.actList.selected
	case tabFollowing:
		total = len(m.following)
		selected = m.followList.selected
	default:
		return
	}
	if total == 0 || m.viewport.Height <= 0 {
		return
	}
	top := m.viewport.YOffset
	bottom := top + m.viewport.Height - 1
	if selected < top {
		m.viewport.YOffset = selected
	} else if selected > bottom {
		m.viewport.YOffset = selected - m.viewport.Height + 1
	}
	if m.viewport.YOffset < 0 {
		m.viewport.YOffset = 0
	}
	maxOffset := max(0, total-m.viewport.Height)
	if m.viewport.YOffset > maxOffset {
		m.viewport.YOffset = maxOffset
	}
}

func (m model) openSelectedProfile() model {
	if m.activeTab != tabFollowing || len(m.following) == 0 {
		return m
	}
	item := m.following[m.followList.selected]
	username := usernameFromURL(item.ActorURL)
	if username == "" {
		username = usernameFromURL(item.FilmURL)
	}
	if username == "" {
		return m
	}
	m.profileUser = username
	m.loading = true
	m.profileModal = true
	m.modalVP.YOffset = 0
	return m
}

func (m model) openSelectedFilm() model {
	var filmURL string
	switch m.activeTab {
	case tabDiary:
		if len(m.diary) == 0 {
			return m
		}
		filmURL = m.diary[m.diaryList.selected].FilmURL
	case tabWatchlist:
		if len(m.watchlist) == 0 {
			return m
		}
		filmURL = m.watchlist[m.watchList.selected].FilmURL
	case tabActivity:
		if len(m.activity) == 0 {
			return m
		}
		filmURL = m.activity[m.actList.selected].FilmURL
	case tabFollowing:
		if len(m.following) == 0 {
			return m
		}
		filmURL = m.following[m.followList.selected].FilmURL
	}
	filmURL = normalizeFilmURL(filmURL)
	if filmURL == "" {
		return m
	}
	m.film = Film{URL: filmURL}
	m.filmReturn = m.activeTab
	m.activeTab = tabFilm
	m.loading = true
	m.viewport.YOffset = 0
	return m
}

func (m model) goBackProfile() model {
	if len(m.profileStack) == 0 {
		return m
	}
	last := m.profileStack[len(m.profileStack)-1]
	m.profileStack = m.profileStack[:len(m.profileStack)-1]
	m.profileUser = last
	m.activeTab = tabProfile
	m.loading = true
	return m
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
	case tabWatchlist:
		body = renderWatchlist(m, theme)
	case tabFilm:
		body = renderFilm(m, theme)
	case tabActivity:
		body = renderActivity(m.activity, m.activityErr, m.actList.selected, theme)
	case tabFollowing:
		body = renderActivity(m.following, m.followErr, m.followList.selected, theme)
	}

	footer := theme.subtle.Render("tab/shift+tab switch • j/k move • pgup/pgdn scroll • enter view • b back • esc close film • o open browser • r refresh • q quit")
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

func (m model) modalOpen() bool {
	return m.activeTab == tabFilm || m.profileModal
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

func renderTabs(m model, theme themeStyles) string {
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

func renderProfile(m model, theme themeStyles) string {
	if m.profileErr != nil {
		return theme.dim.Render("Error: " + m.profileErr.Error())
	}
	if m.loading && len(m.profile.Stats) == 0 && len(m.profile.Favorites) == 0 {
		return theme.dim.Render("Loading profile…")
	}
	var rows []string
	rows = append(rows, theme.subtle.Render(renderBreadcrumbs(m.profileStack, m.profileUser, theme.user)))
	if len(m.profile.Stats) > 0 {
		rows = append(rows, "", theme.subtle.Render("Stats"))
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
			line = strings.Replace(line, m.profileUser, theme.user.Render(m.profileUser), 1)
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

func renderWatchlist(m model, theme themeStyles) string {
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

func renderFilm(m model, theme themeStyles) string {
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

func renderFilmModal(base string, m model, theme themeStyles) string {
	overlay := renderFilm(m, theme)
	if overlay == "" {
		return base
	}
	width, height := modalDimensions(m.width, m.height)
	panel := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3A4A55")).
		Background(lipgloss.Color("#14181C")).
		Foreground(lipgloss.Color("#E6F0F2"))
	panelContent := lipgloss.Place(width-4, height-2, lipgloss.Left, lipgloss.Top, overlay)
	modal := panel.Render(panelContent)

	dim := lipgloss.NewStyle().
		Background(lipgloss.Color("#0E1114")).
		Foreground(lipgloss.Color("#5E6A72")).
		Render(base)
	return dim + "\n" + lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceBackground(lipgloss.Color("#0E1114")))
}

func renderProfileModal(base string, m model, theme themeStyles) string {
	overlay := renderProfile(m, theme)
	if overlay == "" {
		return base
	}
	width, height := modalDimensions(m.width, m.height)
	panel := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3A4A55")).
		Background(lipgloss.Color("#14181C")).
		Foreground(lipgloss.Color("#E6F0F2"))
	panelContent := lipgloss.Place(width-4, height-2, lipgloss.Left, lipgloss.Top, overlay)
	modal := panel.Render(panelContent)

	dim := lipgloss.NewStyle().
		Background(lipgloss.Color("#0E1114")).
		Foreground(lipgloss.Color("#5E6A72")).
		Render(base)
	return dim + "\n" + lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modal, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceBackground(lipgloss.Color("#0E1114")))
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

func renderSummary(item ActivityItem, theme themeStyles) string {
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

func parseWatchlist(doc *goquery.Document) ([]WatchlistItem, error) {
	var items []WatchlistItem
	doc.Find(".js-watchlist-main-content .react-component").Each(func(_ int, item *goquery.Selection) {
		title := strings.TrimSpace(item.AttrOr("data-item-name", ""))
		filmURL := strings.TrimSpace(item.AttrOr("data-item-link", ""))
		if title == "" || filmURL == "" {
			return
		}
		if !strings.Contains(filmURL, "/film/") {
			return
		}
		if strings.HasPrefix(filmURL, "/") {
			filmURL = baseURL + filmURL
		}
		year := ""
		if open := strings.LastIndex(title, "("); open != -1 {
			if close := strings.LastIndex(title, ")"); close > open {
				year = strings.TrimSpace(title[open+1 : close])
				title = strings.TrimSpace(title[:open])
			}
		}
		items = append(items, WatchlistItem{
			Title:   title,
			FilmURL: filmURL,
			Year:    year,
		})
	})
	return items, nil
}

func parseActivity(doc *goquery.Document) ([]ActivityItem, error) {
	var items []ActivityItem
	doc.Find("section.activity-row").Each(func(_ int, row *goquery.Selection) {
		kind := strings.TrimSpace(row.AttrOr("class", ""))
		summarySel := row.Find(".activity-summary").First()
		summary := strings.TrimSpace(summarySel.Text())
		when := strings.TrimSpace(row.Find("time.time").First().AttrOr("datetime", ""))
		actorSel := row.Find(".activity-summary a.name").First()
		if actorSel.Length() == 0 {
			actorSel = row.Find(".attribution-detail a.owner").First()
		}
		actor := strings.TrimSpace(actorSel.Text())
		actorURL, _ := actorSel.Attr("href")
		if actorURL != "" && strings.HasPrefix(actorURL, "/") {
			actorURL = baseURL + actorURL
		}
		targetSel := row.Find(".activity-summary a.target").First()
		title := strings.TrimSpace(targetSel.Text())
		filmURL, _ := targetSel.Attr("href")
		if title == "" {
			titleSel := row.Find("h2.name a").First()
			title = strings.TrimSpace(titleSel.Text())
			filmURL, _ = titleSel.Attr("href")
		}
		if filmURL != "" && strings.HasPrefix(filmURL, "/") {
			filmURL = baseURL + filmURL
		}
		rating := strings.TrimSpace(row.Find(".rating").First().Text())
		parts := parseSummaryParts(summarySel)
		items = append(items, ActivityItem{
			Summary:  summary,
			When:     when,
			Title:    title,
			FilmURL:  filmURL,
			Rating:   rating,
			Kind:     kind,
			Actor:    actor,
			ActorURL: actorURL,
			Parts:    parts,
		})
	})
	return items, nil
}

func parseFilm(doc *goquery.Document, url string) (Film, error) {
	var film Film
	film.URL = url
	title := strings.TrimSpace(doc.Find(`meta[property="og:title"]`).AttrOr("content", ""))
	description := strings.TrimSpace(doc.Find(`meta[property="og:description"]`).AttrOr("content", ""))
	director := strings.TrimSpace(doc.Find(`meta[name="twitter:data1"]`).AttrOr("content", ""))
	avgRating := strings.TrimSpace(doc.Find(`meta[name="twitter:data2"]`).AttrOr("content", ""))
	runtime := strings.TrimSpace(findRuntime(doc))
	cast := parseTopBilledCast(doc, 6)

	year := ""
	if open := strings.LastIndex(title, "("); open != -1 {
		if close := strings.LastIndex(title, ")"); close > open {
			year = strings.TrimSpace(title[open+1 : close])
			title = strings.TrimSpace(title[:open])
		}
	}
	film.Title = title
	film.Year = year
	film.Description = description
	film.Director = director
	film.AvgRating = avgRating
	film.Runtime = runtime
	film.Cast = cast
	return film, nil
}

func findRuntime(doc *goquery.Document) string {
	text := compactSpaces(doc.Find("p.text-link.text-footer").First().Text())
	if text == "" {
		return ""
	}
	fields := strings.Fields(text)
	for i := 0; i < len(fields); i++ {
		if strings.HasSuffix(fields[i], "mins") {
			return fields[i]
		}
		if i+1 < len(fields) && strings.HasSuffix(fields[i+1], "mins") {
			return fields[i] + " " + fields[i+1]
		}
	}
	return ""
}

func parseTopBilledCast(doc *goquery.Document, limit int) []string {
	var cast []string
	doc.Find("#tab-cast .cast-list a.text-slug").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		name := strings.TrimSpace(sel.Text())
		if name == "" {
			return true
		}
		cast = append(cast, name)
		if limit > 0 && len(cast) >= limit {
			return false
		}
		return true
	})
	return cast
}

func parseUserFilm(doc *goquery.Document) (string, string) {
	status := strings.TrimSpace(doc.Find(".film-viewing-info-wrapper .context").First().Text())
	status = strings.TrimSuffix(status, "by")
	status = strings.TrimSpace(strings.TrimSuffix(status, "By"))
	rating := strings.TrimSpace(doc.Find(".content-reactions-strip .rating").First().Text())
	return rating, status
}

func userFilmURL(username, filmURL string) string {
	slug := filmSlug(filmURL)
	if slug == "" || username == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/film/%s/", baseURL, username, slug)
}

func filmSlug(filmURL string) string {
	filmURL = normalizeFilmURL(filmURL)
	if filmURL == "" {
		return ""
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(filmURL, baseURL), "/"), "/")
	if len(parts) >= 2 && parts[0] == "film" {
		return parts[1]
	}
	return ""
}

func parseSummaryParts(summary *goquery.Selection) []SummaryPart {
	var parts []SummaryPart
	if summary == nil {
		return parts
	}
	summary.Contents().Each(func(_ int, node *goquery.Selection) {
		n := node.Get(0)
		if n == nil {
			return
		}
		switch n.Type {
		case html.TextNode:
			addSummaryPart(&parts, n.Data, "text")
		case html.ElementNode:
			if node.Is("a") {
				class := node.AttrOr("class", "")
				text := node.Text()
				switch {
				case strings.Contains(class, "target"):
					if n != nil {
						extracted := extractTargetTitle(n)
						if extracted != "" {
							text = extracted
						}
					}
					addSummaryPart(&parts, text, "movie")
				case strings.Contains(class, "name"):
					addSummaryPart(&parts, text, "user")
				default:
					addSummaryPart(&parts, text, "text")
				}
				return
			}
			if node.Is("span") {
				class := node.AttrOr("class", "")
				text := node.Text()
				if strings.Contains(class, "rating") {
					addSummaryPart(&parts, text, "rating")
				} else {
					addSummaryPart(&parts, text, "text")
				}
				return
			}
			if node.Is("strong") {
				class := node.AttrOr("class", "")
				text := node.Text()
				if strings.Contains(class, "name") {
					addSummaryPart(&parts, text, "user")
				} else {
					addSummaryPart(&parts, text, "text")
				}
				return
			}
			addSummaryPart(&parts, node.Text(), "text")
		}
	})
	return parts
}

func addSummaryPart(parts *[]SummaryPart, text, kind string) {
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	*parts = append(*parts, SummaryPart{Text: text, Kind: kind})
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

func extractTargetTitle(node *html.Node) string {
	var parts []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n == nil {
			return
		}
		switch n.Type {
		case html.TextNode:
			parts = append(parts, n.Data)
		case html.ElementNode:
			if n.Data == "span" {
				class := attrValue(n, "class")
				if strings.Contains(class, "context") || strings.Contains(class, "rating") {
					return
				}
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				walk(c)
			}
		}
	}
	walk(node)
	return compactSpaces(strings.Join(parts, " "))
}

func attrValue(n *html.Node, key string) string {
	for _, attr := range n.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func usernameFromURL(url string) string {
	if url == "" {
		return ""
	}
	if strings.HasPrefix(url, baseURL) {
		url = strings.TrimPrefix(url, baseURL)
	}
	url = strings.TrimSpace(url)
	if !strings.HasPrefix(url, "/") {
		return ""
	}
	parts := strings.Split(strings.Trim(url, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func profileURL(username string) string {
	if username == "" {
		return ""
	}
	return fmt.Sprintf("%s/%s/", baseURL, username)
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

func openBrowserCmd(url string) tea.Cmd {
	return func() tea.Msg {
		if url == "" {
			return openMsg{err: fmt.Errorf("missing profile URL")}
		}
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "darwin":
			cmd = exec.Command("open", url)
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", url)
		default:
			cmd = exec.Command("xdg-open", url)
		}
		if err := cmd.Start(); err != nil {
			return openMsg{err: err}
		}
		return openMsg{}
	}
}

func normalizeFilmURL(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return ""
	}
	if strings.HasPrefix(url, baseURL) {
		url = strings.TrimPrefix(url, baseURL)
	}
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}
	parts := strings.Split(strings.Trim(url, "/"), "/")
	if len(parts) == 0 {
		return ""
	}
	if parts[0] == "film" && len(parts) >= 2 {
		return baseURL + "/film/" + parts[1] + "/"
	}
	for i := len(parts) - 1; i >= 0; i-- {
		if parts[i] == "film" && i+1 < len(parts) {
			return baseURL + "/film/" + parts[i+1] + "/"
		}
	}
	return ""
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
