package cmdpalette

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kienlt/es-cli/internal/tui/commands"
	"github.com/kienlt/es-cli/internal/tui/theme"
)

type SubmitMsg struct {
	Command string
}

type CancelMsg struct{}

type Model struct {
	input  textinput.Model
	router *commands.Router
	ghost  string
	width  int
}

func New(router *commands.Router, width int) Model {
	ti := textinput.New()
	ti.Prompt = ":"
	ti.Focus()
	ti.CharLimit = 64

	return Model{
		input:  ti,
		router: router,
		width:  width,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.Type {
		case tea.KeyEsc:
			return m, func() tea.Msg { return CancelMsg{} }

		case tea.KeyEnter:
			val := m.input.Value()
			if cmd := m.router.Match(val); cmd != nil {
				return m, func() tea.Msg { return SubmitMsg{Command: cmd.Name} }
			}
			// No match — try with ghost completion
			if m.ghost != "" {
				full := val + m.ghost
				if cmd := m.router.Match(full); cmd != nil {
					return m, func() tea.Msg { return SubmitMsg{Command: cmd.Name} }
				}
			}
			return m, nil

		case tea.KeyTab:
			if m.ghost != "" {
				m.input.SetValue(m.input.Value() + m.ghost)
				m.input.CursorEnd()
				m.ghost = ""
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	// Update ghost text
	m.ghost = ""
	val := m.input.Value()
	if val != "" {
		matches := m.router.Complete(val)
		if len(matches) > 0 {
			name := matches[0].Name
			if strings.HasPrefix(name, val) && name != val {
				m.ghost = name[len(val):]
			}
		}
	}

	return m, cmd
}

func (m Model) View() string {
	inputView := m.input.View()

	ghost := ""
	if m.ghost != "" {
		ghost = lipgloss.NewStyle().Foreground(theme.ColorCyan).Faint(true).Render(m.ghost)
	}

	content := " " + inputView + ghost
	return theme.StatusBarStyle.Width(m.width).Render(content)
}
