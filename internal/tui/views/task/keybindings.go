package task

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Up             key.Binding
	Down           key.Binding
	SortByDuration key.Binding
	SortByAction   key.Binding
	SortByNode     key.Binding
	CycleCategory  key.Binding
	ToggleAll      key.Binding
	Cancel         key.Binding
	CancelAll      key.Binding
	Detail         key.Binding
	Refresh        key.Binding
	Search         key.Binding
	Help           key.Binding
	Quit           key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		SortByDuration: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("Shift+D", "sort by duration"),
		),
		SortByAction: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("Shift+A", "sort by action"),
		),
		SortByNode: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("Shift+N", "sort by node"),
		),
		CycleCategory: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "cycle category"),
		),
		ToggleAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "toggle top 20/all"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "cancel task"),
		),
		CancelAll: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("Shift+C", "cancel all tasks"),
		),
		Detail: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "task detail"),
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
