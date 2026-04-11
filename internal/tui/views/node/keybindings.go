package node

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	SortByName  key.Binding
	SortByCPU   key.Binding
	SortByHeap  key.Binding
	SortByRAM   key.Binding
	SortByDisk  key.Binding
	Maintenance key.Binding
	Refresh     key.Binding
	Search      key.Binding
	Help        key.Binding
	Quit        key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		SortByName: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("Shift+N", "sort by name"),
		),
		SortByCPU: key.NewBinding(
			key.WithKeys("C"),
			key.WithHelp("Shift+C", "sort by cpu"),
		),
		SortByHeap: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("Shift+H", "sort by heap"),
		),
		SortByRAM: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("Shift+R", "sort by ram"),
		),
		SortByDisk: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("Shift+D", "sort by disk"),
		),
		Maintenance: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "maintenance"),
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
