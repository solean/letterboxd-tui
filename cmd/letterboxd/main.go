package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"

	"github.com/solean/letterboxd-tui/internal/config"
	"github.com/solean/letterboxd-tui/internal/letterboxd"
	"github.com/solean/letterboxd-tui/internal/logging"
	"github.com/solean/letterboxd-tui/internal/ui"
	"github.com/solean/letterboxd-tui/internal/version"
)

func main() {
	var userFlag string
	var setupFlag bool
	var noCookieFlag bool
	var versionFlag bool
	var debugFlag bool
	flag.StringVar(&userFlag, "user", "", "Letterboxd username")
	flag.BoolVar(&setupFlag, "setup", false, "Run first-time setup")
	flag.BoolVar(&noCookieFlag, "no-cookie", false, "Run without a stored cookie")
	flag.BoolVar(&versionFlag, "version", false, "Print version and exit")
	flag.BoolVar(&debugFlag, "debug", false, "Show debug errors (stack traces, HTTP details)")
	interactive := isInteractiveTTY()
	if !interactive && wantsHelp(os.Args[1:]) {
		flag.Usage()
		return
	}
	flag.Parse()

	if versionFlag {
		fmt.Printf("letterboxd-tui %s\n", version.String())
		return
	}
	if !interactive {
		fmt.Fprintln(os.Stderr, "This app requires a TTY; run in a terminal.")
		flag.Usage()
		os.Exit(2)
	}

	state, err := resolveStartup(strings.TrimSpace(userFlag))
	if err != nil {
		logging.LogError("startup", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if setupFlag {
		state.needUsername = true
		state.needCookie = true
		state.username = ""
		state.cookie = ""
	}
	if noCookieFlag {
		state.cookie = ""
		state.needCookie = false
	}
	if state.needUsername || state.needCookie {
		result, err := ui.RunOnboarding(ui.OnboardingOptions{
			Username:   state.username,
			Cookie:     state.cookie,
			NeedUser:   state.needUsername,
			NeedCookie: state.needCookie,
			ConfigPath: state.configPath,
		})
		if err != nil {
			logging.LogError("onboarding", err)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if result.Cancelled {
			os.Exit(0)
		}
		if state.needUsername {
			state.username = strings.TrimSpace(result.Username)
		}
		if state.needCookie {
			if strings.TrimSpace(result.Cookie) != "" {
				state.cookie = strings.TrimSpace(result.Cookie)
			}
		}
		if state.username != "" {
			state.config.Username = state.username
			state.configDirty = true
		}
		if strings.TrimSpace(result.Cookie) != "" {
			state.config.Cookie = result.Cookie
			state.configDirty = true
		}
	}
	if state.configDirty {
		if err := config.Save(state.config); err != nil {
			logging.LogError("config save", err)
			fmt.Fprintln(os.Stderr, "warning: unable to save config:", err)
		}
	}
	if strings.TrimSpace(state.username) == "" {
		fmt.Fprintln(os.Stderr, "missing Letterboxd username (use -user or set LETTERBOXD_USER)")
		flag.Usage()
		os.Exit(2)
	}

	cookie := state.cookie
	client := letterboxd.NewClient(nil, cookie)
	client.Debug = debugFlag || envBool("LETTERBOXD_DEBUG")

	m := ui.NewModel(state.username, client)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		logging.LogError("bubbletea run", err)
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type startupState struct {
	username     string
	cookie       string
	config       config.Config
	configPath   string
	configDirty  bool
	needUsername bool
	needCookie   bool
}

func resolveStartup(userFlag string) (startupState, error) {
	state := startupState{}
	if path, err := config.Path(); err == nil {
		state.configPath = path
	}

	cfg, err := config.Load()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return startupState{}, err
		}
		cfg = config.Config{}
	}
	state.config = cfg

	envUser := strings.TrimSpace(os.Getenv("LETTERBOXD_USER"))
	username := strings.TrimSpace(userFlag)
	if username == "" {
		if envUser != "" {
			username = envUser
		} else if cfg.Username != "" {
			username = cfg.Username
		}
	}

	cookie := strings.TrimSpace(cfg.Cookie)

	state.username = username
	state.cookie = cookie
	state.needUsername = strings.TrimSpace(username) == ""
	state.needCookie = cookieNeedsPrompt(cookie)
	return state, nil
}

func cookieNeedsPrompt(cookie string) bool {
	cookie = strings.TrimSpace(cookie)
	if cookie == "" {
		return true
	}
	return !strings.Contains(cookie, "com.xk72.webparts.csrf=")
}

func isInteractiveTTY() bool {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}
	termEnv := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
	return termEnv != "" && termEnv != "dumb"
}

func wantsHelp(args []string) bool {
	for _, arg := range args {
		switch arg {
		case "-h", "-help", "--help":
			return true
		}
	}
	return false
}

func envBool(key string) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
