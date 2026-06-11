package threadpool

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	SortByName     key.Binding
	SortByNode     key.Binding
	SortByActive   key.Binding
	SortByQueue    key.Binding
	SortByRejected key.Binding
	ToggleAll      key.Binding
	Refresh        key.Binding
	Search         key.Binding
	Help           key.Binding
	Quit           key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		SortByName: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("Shift+A", "sort by name"),
		),
		SortByNode: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("Shift+N", "sort by node"),
		),
		SortByActive: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("Shift+C", "sort by active"),
		),
		SortByQueue: key.NewBinding(
			key.WithKeys("Q"),
			key.WithHelp("Shift+Q", "sort by queue"),
		),
		SortByRejected: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("Shift+R", "sort by rejected"),
		),
		ToggleAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "toggle direct pools"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
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
