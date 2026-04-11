package createilm

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kienlt/es-cli/internal/tui/theme"
)

type SubmitMsg struct {
	Name    string
	Body    string
	Editing bool
}

type CancelMsg struct{}

type field int

const (
	fieldName field = iota
	fieldDeleteAfter
	fieldCount
)

type Model struct {
	inputs  []textinput.Model
	focused field
	err     string
	editing bool
}

func New() Model {
	return newModel(false, "", "")
}

func NewEdit(name, deleteAfter string) Model {
	return newModel(true, name, deleteAfter)
}

func newModel(editing bool, name, deleteAfter string) Model {
	inputs := make([]textinput.Model, fieldCount)

	nameInput := textinput.New()
	nameInput.Placeholder = "my-policy"
	nameInput.CharLimit = 255
	if name != "" {
		nameInput.SetValue(name)
	}
	inputs[fieldName] = nameInput

	deleteInput := textinput.New()
	deleteInput.Placeholder = "30d"
	deleteInput.CharLimit = 20
	if deleteAfter != "" {
		deleteInput.SetValue(deleteAfter)
	}
	inputs[fieldDeleteAfter] = deleteInput

	focused := fieldName
	if editing {
		// When editing, name is read-only — focus on delete after
		focused = fieldDeleteAfter
		inputs[fieldDeleteAfter].Focus()
	} else {
		inputs[fieldName].Focus()
	}

	return Model{inputs: inputs, focused: focused, editing: editing}
}

func (m Model) Init() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return CancelMsg{} }
		case "tab", "shift+tab":
			if m.editing {
				// When editing, only delete after field is editable
				m.focused = fieldDeleteAfter
			} else {
				m.focused = (m.focused + 1) % fieldCount
			}
			for i := range m.inputs {
				if field(i) == m.focused {
					m.inputs[i].Focus()
				} else {
					m.inputs[i].Blur()
				}
			}
			return m, nil
		case "enter":
			name := strings.TrimSpace(m.inputs[fieldName].Value())
			if name == "" {
				m.err = "Policy name is required"
				return m, nil
			}
			deleteAfter := strings.TrimSpace(m.inputs[fieldDeleteAfter].Value())
			if deleteAfter == "" {
				m.err = "Delete after is required (e.g. 30d)"
				return m, nil
			}
			m.err = ""
			body := m.BuildJSON()
			editing := m.editing
			return m, func() tea.Msg { return SubmitMsg{Name: name, Body: body, Editing: editing} }
		}
	}
	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m Model) BuildJSON() string {
	deleteAfter := strings.TrimSpace(m.inputs[fieldDeleteAfter].Value())
	return fmt.Sprintf(`{"policy":{"phases":{"hot":{"min_age":"0ms","actions":{"rollover":{"max_age":"30d","max_primary_shard_size":"50gb"}}},"delete":{"min_age":%q,"actions":{"delete":{}}}}}}`, deleteAfter)
}

func (m Model) View() string {
	title := "Create ILM Policy"
	if m.editing {
		title = "Edit ILM Policy"
	}

	labelStyle := lipgloss.NewStyle().Width(14).Foreground(theme.ColorWhite).Bold(true)

	var form strings.Builder
	form.WriteString(theme.ModalTitleStyle.Render(title) + "\n\n")

	// Name field
	if m.editing {
		form.WriteString("  " + labelStyle.Render("Name:") + " " + theme.HelpKeyStyle.Render(m.inputs[fieldName].Value()) + "\n")
	} else {
		cursor := "  "
		if m.focused == fieldName {
			cursor = theme.HelpKeyStyle.Render("> ")
		}
		form.WriteString(cursor + labelStyle.Render("Name:") + " " + m.inputs[fieldName].View() + "\n")
	}

	// Delete After field
	cursor := "  "
	if m.focused == fieldDeleteAfter {
		cursor = theme.HelpKeyStyle.Render("> ")
	}
	form.WriteString(cursor + labelStyle.Render("Delete After:") + " " + m.inputs[fieldDeleteAfter].View() + "\n")

	if m.err != "" {
		form.WriteString("\n" + theme.ErrorStyle.Render(m.err))
	}
	form.WriteString("\n" + theme.HelpDescStyle.Render("tab: next • enter: save • esc: cancel"))

	return theme.ModalStyle.Render(form.String())
}
