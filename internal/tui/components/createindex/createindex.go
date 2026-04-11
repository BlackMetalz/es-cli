package createindex

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kienlt/es-cli/internal/tui/theme"
)

type SubmitMsg struct {
	Name     string
	Shards   int
	Replicas int
}

type CancelMsg struct{}

type field int

const (
	fieldName field = iota
	fieldShards
	fieldReplicas
	fieldCount
)

type Model struct {
	inputs  []textinput.Model
	focused field
	err     string
}

func New() Model {
	inputs := make([]textinput.Model, fieldCount)

	nameInput := textinput.New()
	nameInput.Placeholder = "index-name"
	nameInput.Focus()
	nameInput.CharLimit = 255
	inputs[fieldName] = nameInput

	shardsInput := textinput.New()
	shardsInput.Placeholder = "1"
	shardsInput.SetValue("1")
	shardsInput.CharLimit = 5
	inputs[fieldShards] = shardsInput

	replicasInput := textinput.New()
	replicasInput.Placeholder = "1"
	replicasInput.SetValue("1")
	replicasInput.CharLimit = 5
	inputs[fieldReplicas] = replicasInput

	return Model{
		inputs:  inputs,
		focused: fieldName,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg { return CancelMsg{} }

		case "tab", "shift+tab":
			if msg.String() == "tab" {
				m.focused = (m.focused + 1) % fieldCount
			} else {
				m.focused = (m.focused - 1 + fieldCount) % fieldCount
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
				m.err = "Index name cannot be empty"
				return m, nil
			}

			shards, err := strconv.Atoi(m.inputs[fieldShards].Value())
			if err != nil || shards < 1 {
				m.err = "Number of shards must be a positive integer"
				return m, nil
			}

			replicas, err := strconv.Atoi(m.inputs[fieldReplicas].Value())
			if err != nil || replicas < 0 {
				m.err = "Number of replicas must be a non-negative integer"
				return m, nil
			}

			m.err = ""
			return m, func() tea.Msg {
				return SubmitMsg{Name: name, Shards: shards, Replicas: replicas}
			}
		}
	}

	// Update focused input
	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

func (m Model) View() string {
	var b strings.Builder

	title := theme.ModalTitleStyle.Render("Create New Index")
	b.WriteString(title + "\n\n")

	labels := []string{"Index Name:", "Shards:    ", "Replicas:  "}
	for i, label := range labels {
		cursor := "  "
		if field(i) == m.focused {
			cursor = "> "
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", cursor, label, m.inputs[i].View()))
	}

	if m.err != "" {
		b.WriteString("\n" + theme.ErrorStyle.Render(m.err))
	}

	b.WriteString("\n" + theme.HelpDescStyle.Render("tab: next field • enter: create • esc: cancel"))

	return theme.ModalStyle.Render(b.String())
}
