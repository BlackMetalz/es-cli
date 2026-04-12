package query

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Search    key.Binding
	RawQuery  key.Binding
	Follow    key.Binding
	TimeRange key.Binding
	Columns   key.Binding
	Refresh   key.Binding
	NextPage  key.Binding
	PrevPage  key.Binding
	Expand    key.Binding
	Back      key.Binding
	Help      key.Binding
	Quit      key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "query builder"),
		),
		RawQuery: key.NewBinding(
			key.WithKeys("!"),
			key.WithHelp("!", "raw query"),
		),
		Follow: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "follow"),
		),
		TimeRange: key.NewBinding(
			key.WithKeys("t"),
			key.WithHelp("t", "time range"),
		),
		Columns: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "columns"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]", "page down"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[", "page up"),
		),
		Expand: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "expand doc"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
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
