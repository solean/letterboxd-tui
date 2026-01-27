package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type OnboardingOptions struct {
	Username   string
	Cookie     string
	NeedUser   bool
	NeedCookie bool
	ConfigPath string
}

type OnboardingResult struct {
	Username  string
	Cookie    string
	Cancelled bool
}

func RunOnboarding(opts OnboardingOptions) (OnboardingResult, error) {
	model := newOnboardingModel(opts)
	p := tea.NewProgram(model, tea.WithAltScreen())
	final, err := p.Run()
	if err != nil {
		return OnboardingResult{}, err
	}
	if m, ok := final.(onboardingModel); ok {
		return m.result, nil
	}
	return OnboardingResult{}, fmt.Errorf("unexpected onboarding model")
}

type onboardingStage int

const (
	stageSplash onboardingStage = iota
	stageUsername
	stageCookie
	stageDone
)

type onboardingModel struct {
	stage      onboardingStage
	steps      []onboardingStage
	width      int
	height     int
	styles     onboardingStyles
	spinner    spinner.Model
	username   textinput.Model
	cookie     textinput.Model
	configPath string
	errorMsg   string
	cookieHelp bool
	cookieWarn string
	result     OnboardingResult
}

type onboardingStyles struct {
	bg         lipgloss.Style
	panel      lipgloss.Style
	title      lipgloss.Style
	subtitle   lipgloss.Style
	accent     lipgloss.Style
	dim        lipgloss.Style
	step       lipgloss.Style
	label      lipgloss.Style
	input      lipgloss.Style
	inputFocus lipgloss.Style
	warning    lipgloss.Style
	footer     lipgloss.Style
}

func newOnboardingStyles() onboardingStyles {
	bg := lipgloss.Color("#0E1114")
	fg := lipgloss.Color("#E6F0F2")
	accent := lipgloss.Color("#00E054")
	orange := lipgloss.Color("#FF8C3A")
	panelBG := lipgloss.Color("#14181C")
	panelBorder := lipgloss.Color("#3A4A55")
	return onboardingStyles{
		bg:         lipgloss.NewStyle().Background(bg).Foreground(fg),
		panel:      lipgloss.NewStyle().Background(panelBG).Foreground(fg).Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(panelBorder),
		title:      lipgloss.NewStyle().Foreground(accent).Bold(true),
		subtitle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#9BB0B8")),
		accent:     lipgloss.NewStyle().Foreground(accent).Bold(true),
		dim:        lipgloss.NewStyle().Foreground(lipgloss.Color("#7F8D96")),
		step:       lipgloss.NewStyle().Foreground(lipgloss.Color("#9BB0B8")).Bold(true),
		label:      lipgloss.NewStyle().Foreground(fg).Bold(true),
		input:      lipgloss.NewStyle().Foreground(fg),
		inputFocus: lipgloss.NewStyle().Foreground(fg).Background(lipgloss.Color("#1F2A33")),
		warning:    lipgloss.NewStyle().Foreground(orange).Bold(true),
		footer:     lipgloss.NewStyle().Foreground(lipgloss.Color("#9BB0B8")),
	}
}

func newOnboardingModel(opts OnboardingOptions) onboardingModel {
	styles := newOnboardingStyles()
	spin := spinner.New(spinner.WithSpinner(spinner.Line))
	spin.Style = styles.accent

	userInput := textinput.New()
	userInput.Placeholder = "your-username"
	userInput.CharLimit = 40
	userInput.SetValue(strings.TrimSpace(opts.Username))

	cookieInput := textinput.New()
	cookieInput.Placeholder = "com.xk72.webparts.csrf=..."
	cookieInput.CharLimit = 0
	cookieInput.EchoMode = textinput.EchoPassword
	cookieInput.EchoCharacter = '*'
	cookieInput.SetValue(strings.TrimSpace(opts.Cookie))

	steps := make([]onboardingStage, 0, 2)
	if opts.NeedUser {
		steps = append(steps, stageUsername)
	}
	if opts.NeedCookie {
		steps = append(steps, stageCookie)
	}

	model := onboardingModel{
		stage:      stageSplash,
		steps:      steps,
		styles:     styles,
		spinner:    spin,
		username:   userInput,
		cookie:     cookieInput,
		configPath: strings.TrimSpace(opts.ConfigPath),
	}
	if strings.TrimSpace(opts.Cookie) != "" && !cookieHasCSRF(opts.Cookie) {
		model.cookieWarn = "Cookie looks incomplete; include com.xk72.webparts.csrf=..."
	}
	return model
}

func (m onboardingModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m onboardingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		inputWidth := max(20, min(56, m.width-12))
		m.username.Width = inputWidth
		m.cookie.Width = inputWidth
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.result.Cancelled = true
			return m, tea.Quit
		}
	}

	switch m.stage {
	case stageSplash:
		return m.updateSplash(msg)
	case stageUsername:
		return m.updateUsername(msg)
	case stageCookie:
		return m.updateCookie(msg)
	case stageDone:
		return m.updateDone(msg)
	default:
		return m, nil
	}
}

func (m onboardingModel) View() string {
	bg := m.styles.bg
	switch m.stage {
	case stageSplash:
		return bg.Render(m.renderSplash())
	case stageUsername:
		return bg.Render(m.renderUsername())
	case stageCookie:
		return bg.Render(m.renderCookie())
	case stageDone:
		return bg.Render(m.renderDone())
	default:
		return bg.Render("")
	}
}

func (m onboardingModel) updateSplash(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter", " ":
			if len(m.steps) == 0 {
				m.setStage(stageDone)
				return m, nil
			}
			m.setStage(m.steps[0])
			return m, nil
		}
	}
	return m, nil
}

func (m onboardingModel) updateUsername(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			value := strings.TrimSpace(m.username.Value())
			if value == "" {
				m.errorMsg = "Please enter your Letterboxd username."
				return m, nil
			}
			m.result.Username = value
			return m.advance(), nil
		case "esc":
			m.setStage(stageSplash)
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.username, cmd = m.username.Update(msg)
	return m, cmd
}

func (m onboardingModel) updateCookie(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			value := strings.TrimSpace(m.cookie.Value())
			if value != "" && !cookieHasCSRF(value) {
				m.errorMsg = "Cookie missing com.xk72.webparts.csrf. Paste the full cookie string."
				return m, nil
			}
			m.result.Cookie = value
			return m.advance(), nil
		case "esc":
			if m.cookieHelp {
				m.cookieHelp = false
				return m, nil
			}
			m.result.Cookie = ""
			return m.advance(), nil
		case "shift+tab", "up":
			if m.hasStep(stageUsername) {
				m.setStage(stageUsername)
				return m, nil
			}
		case "?", "h":
			m.cookieHelp = !m.cookieHelp
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.cookie, cmd = m.cookie.Update(msg)
	return m, cmd
}

func (m onboardingModel) updateDone(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter", "q", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m onboardingModel) advance() onboardingModel {
	m.errorMsg = ""
	if len(m.steps) == 0 {
		m.setStage(stageDone)
		return m
	}
	current := m.stage
	for i, step := range m.steps {
		if step == current {
			if i+1 < len(m.steps) {
				m.setStage(m.steps[i+1])
				return m
			}
			m.setStage(stageDone)
			return m
		}
	}
	m.setStage(stageDone)
	return m
}

func (m *onboardingModel) setStage(stage onboardingStage) {
	m.stage = stage
	m.errorMsg = ""
	m.username.Blur()
	m.cookie.Blur()
	switch stage {
	case stageUsername:
		m.username.Focus()
	case stageCookie:
		m.cookie.Focus()
	}
}

func (m onboardingModel) renderSplash() string {
	styles := m.styles
	width := max(40, m.width)

	lines := []string{}
	if width >= 62 {
		lines = append(lines,
			"...........................................................................................................",
			"...........................................................................................................",
			"......:-==-:.:-===-:.:-==-:......%@*............:%@...#@:...............%@@:.......................=@@-....",
			"....:********#######*######*:....@@%....:+%@@#-=@@@@@@@@@%:=#@@@+:-@%+@+%@@#@@*::=%@@%+=%@+:#@*-%@%#@@-....",
			"....+*******@%#####%@%######*....@@%....%@#--@@*=@@--=@@--+@@=-+@#=@@%-:%@@--@@#+@@=-@@*:@@@@+-@@#-*@@-....",
			"....=*******@%#####%@%######=....@@%::::@@%**##+-@@::-%@-:+@@***#+=@@=..#@@::%@#*@*:.%@#:%@@@=-@@=.+@@-....",
			".....-*****+-+#####=-*####*-.....@@@@@@=-%@@@@*::#@@#.*@@@:*@@@@@=-@@=..#@@@@@%:-%@@@@#=@@#-@@*+@@@%@@-....",
			".......::::...::::....::::.......::::::..:::::....:::..:::..::::..:::...::::::....::::.:::..:::.::::::.....",
			"...........................................................................................................",
			"...........................................................................................................",
		)
	} else {
		lines = append(lines, "LETTERBOXD TUI")
	}

	title := styles.title.Render(strings.Join(lines, "\n"))
	tagline := styles.subtitle.Render("Letterboxd in your terminal")
	glow := styles.accent.Render("Press Enter to begin")
	spin := styles.dim.Render(m.spinner.View() + " loading Letterboxd")

	content := lipgloss.JoinVertical(lipgloss.Center, title, "", tagline, "", spin, "", glow)
	return m.place(content)
}

func (m onboardingModel) renderUsername() string {
	styles := m.styles
	step := styles.step.Render(m.stepLabel())
	title := styles.title.Render("First, your Letterboxd username")
	sub := styles.subtitle.Render("We use this to load your profile, diary, and watchlist.")
	line := fmt.Sprintf("%s %s", styles.label.Render("Username"), m.username.View())
	input := styles.input.Render(line)
	if m.username.Focused() {
		input = styles.inputFocus.Render(line)
	}
	helper := styles.dim.Render("Example: \"karsten\" from letterboxd.com/karsten/")
	footer := styles.footer.Render("enter continue • esc back")

	return m.renderPanel(step, title, sub, input, helper, footer)
}

func (m onboardingModel) renderCookie() string {
	if m.cookieHelp {
		return m.renderCookieHelp()
	}
	styles := m.styles
	step := styles.step.Render(m.stepLabel())
	title := styles.title.Render("Add your Letterboxd cookie (optional)")
	sub := styles.subtitle.Render("Needed for watchlist changes and logging diary entries.")
	line := fmt.Sprintf("%s %s", styles.label.Render("Cookie"), m.cookie.View())
	input := styles.input.Render(line)
	if m.cookie.Focused() {
		input = styles.inputFocus.Render(line)
	}
	helper := styles.dim.Render("Paste the full cookie string from your browser.")
	if m.configPath != "" {
		path := truncate(m.configPath, max(20, m.width-12))
		helper = lipgloss.JoinVertical(lipgloss.Left, helper, styles.dim.Render("Saved to: "+path))
	}
	if m.cookieWarn != "" {
		helper = lipgloss.JoinVertical(lipgloss.Left, helper, styles.warning.Render(m.cookieWarn))
	}
	footer := styles.footer.Render("enter continue • esc skip • shift+tab back • ? help")

	return m.renderPanel(step, title, sub, input, helper, footer)
}

func (m onboardingModel) renderDone() string {
	styles := m.styles
	title := styles.title.Render("You're set.")
	sub := styles.subtitle.Render("Press Enter to launch Letterboxd TUI.")
	cta := styles.accent.Render("Enter to continue")
	content := lipgloss.JoinVertical(lipgloss.Center, title, "", sub, "", cta)
	return m.place(styles.panel.Render(content))
}

func (m onboardingModel) renderCookieHelp() string {
	styles := m.styles
	title := styles.title.Render("How to find your Letterboxd cookie")
	lines := []string{
		"1) Sign in at letterboxd.com in your browser.",
		"2) Open Developer Tools (Cmd+Option+I on macOS, Ctrl+Shift+I on Windows/Linux).",
		"3) Open the Network tab and refresh the page.",
		"4) Click any request to letterboxd.com, then copy the Cookie request header.",
		"5) Paste it here. Keep it private.",
	}
	important := styles.warning.Render("Make sure the cookie includes com.xk72.webparts.csrf=...")
	footer := styles.footer.Render("esc back • ? back")
	body := strings.Join(lines, "\n")
	if m.configPath != "" {
		path := truncate(m.configPath, max(20, m.width-12))
		body = body + "\n\nSaved to: " + path
	}
	return m.renderPanel(title, styles.dim.Render(body), important, footer)
}

func (m onboardingModel) renderPanel(lines ...string) string {
	styles := m.styles
	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	if m.errorMsg != "" {
		content = lipgloss.JoinVertical(lipgloss.Left, content, "", styles.warning.Render(m.errorMsg))
	}
	panel := styles.panel.Render(content)
	return m.place(panel)
}

func (m onboardingModel) place(content string) string {
	width := max(0, m.width)
	height := max(0, m.height)
	bgColor := lipgloss.Color("#0E1114")
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content, lipgloss.WithWhitespaceChars(" "), lipgloss.WithWhitespaceBackground(bgColor))
}

func (m onboardingModel) stepLabel() string {
	if len(m.steps) == 0 {
		return ""
	}
	current := 0
	for i, step := range m.steps {
		if step == m.stage {
			current = i + 1
			break
		}
	}
	if current == 0 {
		return ""
	}
	return fmt.Sprintf("Step %d of %d", current, len(m.steps))
}

func (m onboardingModel) hasStep(stage onboardingStage) bool {
	for _, step := range m.steps {
		if step == stage {
			return true
		}
	}
	return false
}

func cookieHasCSRF(cookie string) bool {
	return strings.Contains(cookie, "com.xk72.webparts.csrf=")
}
