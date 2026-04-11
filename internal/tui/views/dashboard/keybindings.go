package dashboard

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Refresh key.Binding
	Help    key.Binding
	Quit    key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}
