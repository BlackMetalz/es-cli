package ilm

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	SortByName key.Binding
	Delete     key.Binding
	Create     key.Binding
	Edit       key.Binding
	ViewDetail key.Binding
	Refresh    key.Binding
	ToggleAll  key.Binding
	Search     key.Binding
	Help       key.Binding
	Quit       key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		SortByName: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("Shift+N", "sort by name"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete policy"),
		),
		Create: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new policy"),
		),
		Edit: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "edit policy"),
		),
		ViewDetail: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "view detail"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		ToggleAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "toggle hidden"),
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
