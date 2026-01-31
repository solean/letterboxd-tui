package ui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit            key.Binding
	QuitAll         key.Binding
	NextTab         key.Binding
	PrevTab         key.Binding
	Down            key.Binding
	Up              key.Binding
	PageDown        key.Binding
	PageUp          key.Binding
	Refresh         key.Binding
	Select          key.Binding
	Back            key.Binding
	ModalBack       key.Binding
	JumpTop         key.Binding
	JumpBottom      key.Binding
	Help            key.Binding
	Open            key.Binding
	Log             key.Binding
	WatchlistAdd    key.Binding
	WatchlistRemove key.Binding
	SearchTab       key.Binding
	Cancel          key.Binding
	Submit          key.Binding
	Toggle          key.Binding
	Sort            key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "quit"),
		),
		QuitAll: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
		NextTab: key.NewBinding(
			key.WithKeys("tab", "right"),
			key.WithHelp("tab/right", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab", "left"),
			key.WithHelp("shift+tab/left", "prev tab"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "down"),
		),
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("ctrl+f", "page down"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("ctrl+b"),
			key.WithHelp("ctrl+b", "page up"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		Back: key.NewBinding(
			key.WithKeys("b"),
			key.WithHelp("b", "back"),
		),
		ModalBack: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "back"),
		),
		JumpTop: key.NewBinding(
			key.WithKeys("g"),
			key.WithHelp("gg", "top"),
		),
		JumpBottom: key.NewBinding(
			key.WithKeys("G"),
			key.WithHelp("G", "bottom"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open in browser"),
		),
		Log: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "log entry"),
		),
		WatchlistAdd: key.NewBinding(
			key.WithKeys("w"),
			key.WithHelp("w", "add to watchlist"),
		),
		WatchlistRemove: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "remove from watchlist"),
		),
		SearchTab: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Submit: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "submit"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "toggle"),
		),
		Sort: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "sort"),
		),
	}
}
