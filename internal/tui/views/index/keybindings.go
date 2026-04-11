package index

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	SortByName      key.Binding
	SortByHealth    key.Binding
	SortBySize      key.Binding
	SortByDocs      key.Binding
	ToggleOpenClose key.Binding
	DeleteIndex     key.Binding
	CreateIndex     key.Binding
	Refresh         key.Binding
	ToggleAll       key.Binding
	ViewDetail      key.Binding
	Search          key.Binding
	Help            key.Binding
	Quit            key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		SortByName: key.NewBinding(
			key.WithKeys("I"),
			key.WithHelp("Shift+I", "sort by index"),
		),
		SortByHealth: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("Shift+H", "sort by health"),
		),
		SortBySize: key.NewBinding(
			key.WithKeys("S"),
			key.WithHelp("Shift+S", "sort by size"),
		),
		SortByDocs: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("Shift+C", "sort by docs"),
		),
		ToggleOpenClose: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "open/close index"),
		),
		DeleteIndex: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete index"),
		),
		CreateIndex: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new index"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		ToggleAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "toggle hidden"),
		),
		ViewDetail: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view detail"),
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
