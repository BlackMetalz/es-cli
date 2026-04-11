package allocationmenu

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kienlt/es-cli/internal/tui/theme"
)

type SubmitMsg struct {
	Value string // "primaries", "none", or "" (reset to all)
}

type CancelMsg struct{}

type choice struct {
	label string
	value string
}

var choices = []choice{
	{"all (reset)", ""},
	{"primaries", "primaries"},
	{"none", "none"},
}

type Model struct {
	cursor  int
	current string // current allocation setting
}

func New(currentSetting string) Model {
	return Model{current: currentSetting}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "j", "down":
		m.cursor = (m.cursor + 1) % len(choices)
	case "k", "up":
		m.cursor = (m.cursor - 1 + len(choices)) % len(choices)
	case "enter":
		val := choices[m.cursor].value
		return m, func() tea.Msg { return SubmitMsg{Value: val} }
	case "esc":
		return m, func() tea.Msg { return CancelMsg{} }
	}
	return m, nil
}

func (m Model) View() string {
	var content string
	content += theme.ModalTitleStyle.Render("Cluster Routing Allocation") + "\n\n"

	for i, c := range choices {
		cursor := "  "
		if i == m.cursor {
			cursor = theme.HelpKeyStyle.Render("> ")
		}
		label := c.label
		if i == m.cursor {
			label = theme.HelpKeyStyle.Render(label)
		}
		content += cursor + label + "\n"
	}

	current := "all (default)"
	if m.current != "" {
		current = theme.HealthYellowStyle.Render(m.current)
	}
	content += "\n" + theme.HelpDescStyle.Render("Current: ") + current + "\n\n"
	content += theme.HelpDescStyle.Render("enter: select • esc: cancel")

	return theme.ModalStyle.Render(content)
}
