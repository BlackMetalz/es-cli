package shard

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	SortByIndex key.Binding
	SortByShard key.Binding
	SortByState key.Binding
	SortByNode  key.Binding
	SortByDocs  key.Binding
	SortByStore key.Binding
	Explain     key.Binding
	ToggleAll   key.Binding
	Refresh     key.Binding
	Search      key.Binding
	Help        key.Binding
	Quit        key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		SortByIndex: key.NewBinding(
			key.WithKeys("I"),
			key.WithHelp("Shift+I", "sort by index"),
		),
		SortByShard: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("Shift+S", "sort by shard"),
		),
		SortByState: key.NewBinding(
			key.WithKeys("T"),
			key.WithHelp("Shift+T", "sort by state"),
		),
		SortByNode: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("Shift+N", "sort by node"),
		),
		SortByDocs: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("Shift+D", "sort by docs"),
		),
		SortByStore: key.NewBinding(
			key.WithKeys("O"),
			key.WithHelp("Shift+O", "sort by store"),
		),
		Explain: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "explain alloc"),
		),
		ToggleAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "toggle hidden"),
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
