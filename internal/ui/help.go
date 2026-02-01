package ui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

type helpKeyMap struct {
	short []key.Binding
	full  [][]key.Binding
}

func (h helpKeyMap) ShortHelp() []key.Binding {
	return h.short
}

func (h helpKeyMap) FullHelp() [][]key.Binding {
	return h.full
}

func newHelpKeyMap(short []key.Binding) helpKeyMap {
	return helpKeyMap{short: short, full: [][]key.Binding{short}}
}

func helpBinding(binding key.Binding, helpKey, desc string) key.Binding {
	binding.SetHelp(helpKey, desc)
	return binding
}

func tabHelp(desc string) key.Binding {
	return key.NewBinding(
		key.WithKeys("tab", "shift+tab"),
		key.WithHelp("tab/shift+tab", desc),
	)
}

func navHelp(desc string) key.Binding {
	return key.NewBinding(
		key.WithKeys("j", "k", "down", "up"),
		key.WithHelp("j/k", desc),
	)
}

func pageHelp(desc string) key.Binding {
	return key.NewBinding(
		key.WithKeys("ctrl+b", "ctrl+f"),
		key.WithHelp("ctrl+b/ctrl+f", desc),
	)
}

func backHelp(helpKey string, keys ...string) key.Binding {
	return key.NewBinding(
		key.WithKeys(keys...),
		key.WithHelp(helpKey, "back"),
	)
}

func (m Model) helpMap() helpKeyMap {
	keys := m.keys
	navScroll := navHelp("scroll")
	navMove := navHelp("move")
	page := pageHelp("page")
	helpToggle := keys.Help
	if m.help.ShowAll {
		helpToggle = helpBinding(keys.Help, "?", "hide help")
	}

	switch {
	case m.logModal:
		tabFields := tabHelp("next/prev field")
		enter := helpBinding(keys.Select, "enter", "toggle/submit")
		toggle := keys.Toggle
		submit := keys.Submit
		back := backHelp("esc/q", "esc", "q")
		return newHelpKeyMap([]key.Binding{tabFields, enter, toggle, submit, back, helpToggle, keys.QuitAll})
	case m.profileModal:
		back := backHelp("b/esc/q", "b", "esc", "q")
		nav := navScroll
		short := []key.Binding{nav, page}
		if m.modalProfileSelectableCount() > 0 {
			nav = navMove
			short = []key.Binding{nav, page, helpBinding(keys.Select, "enter", "view film")}
		}
		short = append(short, keys.JumpTop, keys.JumpBottom, keys.Open, back, helpToggle, keys.QuitAll)
		return newHelpKeyMap(short)
	case m.activeTab == tabFilm:
		watchHint := keys.WatchlistAdd
		if inWatchlist, ok := m.watchlistState(); ok && inWatchlist {
			watchHint = keys.WatchlistRemove
		}
		back := tabHelp("back")
		modalBack := backHelp("esc/q", "esc", "q")
		short := []key.Binding{navScroll, page, keys.JumpTop, keys.JumpBottom}
		if m.hasCookie() {
			short = append(short, keys.Log, watchHint)
		}
		short = append(short, keys.Open, back, modalBack, helpToggle, keys.QuitAll)
		return newHelpKeyMap(short)
	case m.activeTab == tabSearch:
		switchTabs := tabHelp("switch tab")
		if m.searchFocusInput {
			enter := helpBinding(keys.Select, "enter", "search")
			escape := helpBinding(keys.Cancel, "esc", "results")
			return newHelpKeyMap([]key.Binding{enter, escape, switchTabs, helpToggle, keys.Quit, keys.QuitAll})
		}
		enter := helpBinding(keys.Select, "enter", "view")
		search := helpBinding(keys.SearchTab, "/", "edit query")
		return newHelpKeyMap([]key.Binding{navMove, page, keys.JumpTop, keys.JumpBottom, enter, search, switchTabs, helpToggle, keys.Quit, keys.QuitAll})
	default:
		switchTabs := tabHelp("switch tab")
		switch m.activeTab {
		case tabProfile:
			nav := navScroll
			if m.profileSelectableCount() > 0 {
				nav = navMove
			}
			short := []key.Binding{nav, page}
			if m.profileSelectableCount() > 0 {
				short = append(short, helpBinding(keys.Select, "enter", "view film"))
			}
			if len(m.profileStack) > 0 {
				short = append(short, keys.Back)
			}
			short = append(short, keys.JumpTop, keys.JumpBottom, keys.Open, keys.SearchTab, switchTabs, keys.Refresh, helpToggle, keys.Quit, keys.QuitAll)
			return newHelpKeyMap(short)
		case tabDiary, tabWatchlist, tabActivity, tabFollowing:
			enter := keys.Select
			if m.activeTab == tabFollowing {
				enter = helpBinding(keys.Select, "enter", "view profile")
			} else {
				enter = helpBinding(keys.Select, "enter", "view film")
			}
			short := []key.Binding{navMove, page, keys.JumpTop, keys.JumpBottom, enter}
			if m.activeTab == tabDiary {
				short = append(short, helpBinding(keys.Sort, "s", "sort: "+m.diarySortLabel()))
			}
			if m.activeTab == tabWatchlist {
				short = append(short, helpBinding(keys.Sort, "s", "sort: "+m.watchlistSortLabel()))
			}
			short = append(short, keys.SearchTab, switchTabs, keys.Refresh, helpToggle, keys.Quit, keys.QuitAll)
			return newHelpKeyMap(short)
		default:
			return newHelpKeyMap([]key.Binding{keys.JumpTop, keys.JumpBottom, keys.SearchTab, switchTabs, keys.Refresh, helpToggle, keys.Quit, keys.QuitAll})
		}
	}
}

func renderHelp(m Model, theme themeStyles, width int) string {
	helper := m.help
	helper.Width = width

	keyStyle := theme.subtle.Copy().Bold(true)
	descStyle := theme.subtle
	sepStyle := theme.subtle
	helper.Styles.ShortKey = keyStyle
	helper.Styles.ShortDesc = descStyle
	helper.Styles.ShortSeparator = sepStyle
	helper.Styles.FullKey = keyStyle
	helper.Styles.FullDesc = descStyle
	helper.Styles.FullSeparator = sepStyle
	helper.Styles.Ellipsis = sepStyle

	return helper.View(m.helpMap())
}

var _ help.KeyMap = helpKeyMap{}
