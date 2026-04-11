package createtemplate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
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
	fieldPatterns
	fieldShards
	fieldReplicas
	fieldILMPolicy
	fieldCount
)

// ExistingTemplate holds info about an existing template for duplicate detection.
type ExistingTemplate struct {
	Name     string
	Patterns []string // e.g. ["logs-*", "metrics-*"]
}

type Model struct {
	inputs      []textinput.Model
	focused     field
	err         string
	warnings    []string
	editing     bool
	ilmPolicies []string
	existing    []ExistingTemplate
}

func New(ilmPolicies []string, existing []ExistingTemplate) Model {
	inputs := make([]textinput.Model, fieldCount)

	nameInput := textinput.New()
	nameInput.Placeholder = "my-template"
	nameInput.Focus()
	nameInput.CharLimit = 255
	inputs[fieldName] = nameInput

	patternsInput := textinput.New()
	patternsInput.Placeholder = "logs-*,metrics-*"
	patternsInput.CharLimit = 512
	inputs[fieldPatterns] = patternsInput

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

	ilmInput := textinput.New()
	ilmInput.Placeholder = "(optional)"
	ilmInput.CharLimit = 255
	inputs[fieldILMPolicy] = ilmInput

	return Model{
		inputs:      inputs,
		focused:     fieldName,
		ilmPolicies: ilmPolicies,
		existing:    existing,
	}
}

// NewEdit creates a form pre-filled with existing template values.
func NewEdit(name, patterns, shards, replicas, ilmPolicy string, ilmPolicies []string, existing []ExistingTemplate) Model {
	m := New(ilmPolicies, existing)
	m.editing = true
	m.inputs[fieldName].SetValue(name)
	if patterns != "" {
		m.inputs[fieldPatterns].SetValue(patterns)
	}
	if shards != "" {
		m.inputs[fieldShards].SetValue(shards)
	}
	if replicas != "" {
		m.inputs[fieldReplicas].SetValue(replicas)
	}
	if ilmPolicy != "" {
		m.inputs[fieldILMPolicy].SetValue(ilmPolicy)
	}
	// Name is read-only when editing
	m.focused = fieldPatterns
	m.inputs[fieldName].Blur()
	m.inputs[fieldPatterns].Focus()
	return m
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
			// If on ILM field with Tab, try autocomplete first
			if msg.String() == "tab" && m.focused == fieldILMPolicy {
				matches := m.matchingPolicies()
				if len(matches) == 1 {
					m.inputs[fieldILMPolicy].SetValue(matches[0])
					m.inputs[fieldILMPolicy].CursorEnd()
					return m, nil
				}
			}

			if msg.String() == "tab" {
				m.focused = (m.focused + 1) % fieldCount
			} else {
				m.focused = (m.focused - 1 + fieldCount) % fieldCount
			}
			// Skip name field when editing
			if m.editing && m.focused == fieldName {
				if msg.String() == "tab" {
					m.focused = fieldPatterns
				} else {
					m.focused = fieldILMPolicy
				}
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
			m.checkWarnings()
			if len(m.warnings) > 0 {
				m.err = "Fix warnings before submitting"
				return m, nil
			}

			name := strings.TrimSpace(m.inputs[fieldName].Value())
			if name == "" {
				m.err = "Template name cannot be empty"
				return m, nil
			}

			patterns := strings.TrimSpace(m.inputs[fieldPatterns].Value())
			if patterns == "" {
				m.err = "Index patterns cannot be empty"
				return m, nil
			}

			body, err := m.buildJSON()
			if err != "" {
				m.err = err
				return m, nil
			}

			m.err = ""
			editing := m.editing
			return m, func() tea.Msg {
				return SubmitMsg{Name: name, Body: body, Editing: editing}
			}
		}
	}

	// Update focused input
	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	m.checkWarnings()
	if len(m.warnings) == 0 {
		m.err = ""
	}
	return m, cmd
}

func (m Model) buildJSON() (string, string) {
	patterns := strings.TrimSpace(m.inputs[fieldPatterns].Value())
	if patterns == "" {
		return "", "Index patterns cannot be empty"
	}

	// Split patterns by comma
	parts := strings.Split(patterns, ",")
	var patternList []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			patternList = append(patternList, p)
		}
	}
	if len(patternList) == 0 {
		return "", "Index patterns cannot be empty"
	}

	// Parse shards
	shardsStr := strings.TrimSpace(m.inputs[fieldShards].Value())
	if shardsStr == "" {
		shardsStr = "1"
	}
	shards, err := strconv.Atoi(shardsStr)
	if err != nil || shards < 1 {
		return "", "Number of shards must be a positive integer"
	}

	// Parse replicas
	replicasStr := strings.TrimSpace(m.inputs[fieldReplicas].Value())
	if replicasStr == "" {
		replicasStr = "1"
	}
	replicas, err := strconv.Atoi(replicasStr)
	if err != nil || replicas < 0 {
		return "", "Number of replicas must be a non-negative integer"
	}

	ilmPolicy := strings.TrimSpace(m.inputs[fieldILMPolicy].Value())

	// Build JSON
	var patternsJSON strings.Builder
	patternsJSON.WriteString("[")
	for i, p := range patternList {
		if i > 0 {
			patternsJSON.WriteString(", ")
		}
		patternsJSON.WriteString(fmt.Sprintf("%q", p))
	}
	patternsJSON.WriteString("]")

	var settings strings.Builder
	settings.WriteString(fmt.Sprintf(`"number_of_shards": %d`, shards))
	settings.WriteString(fmt.Sprintf(`, "number_of_replicas": %d`, replicas))
	if ilmPolicy != "" {
		settings.WriteString(fmt.Sprintf(`, "index.lifecycle.name": %q`, ilmPolicy))
	}

	json := fmt.Sprintf(`{
  "index_patterns": %s,
  "priority": 100,
  "template": {
    "settings": {
      %s
    }
  }
}`, patternsJSON.String(), settings.String())

	return json, ""
}

func (m Model) View() string {
	var b strings.Builder

	titleText := "Create Index Template"
	if m.editing {
		titleText = "Edit Index Template"
	}
	b.WriteString(theme.ModalTitleStyle.Render(titleText) + "\n\n")

	labelStyle := lipgloss.NewStyle().Width(12).Foreground(theme.ColorWhite).Bold(true)
	labels := []string{"Name", "Patterns", "Shards", "Replicas", "ILM Policy"}

	// Left pane: form fields
	var leftPane strings.Builder
	for i, label := range labels {
		lbl := labelStyle.Render(label + ":")
		if m.editing && field(i) == fieldName {
			// Read-only name when editing
			leftPane.WriteString("  " + lbl + " " + theme.HelpKeyStyle.Render(m.inputs[i].Value()) + "\n")
			continue
		}
		cursor := "  "
		if field(i) == m.focused {
			cursor = theme.HelpKeyStyle.Render("> ")
		}
		leftPane.WriteString(cursor + lbl + " " + m.inputs[i].View() + "\n")
	}

	// Show ILM suggestions when focused on ILM field
	if m.focused == fieldILMPolicy && len(m.ilmPolicies) > 0 {
		matches := m.matchingPolicies()
		if len(matches) > 0 && len(matches) <= 5 {
			leftPane.WriteString(theme.HelpDescStyle.Render("  suggestions: "))
			for i, p := range matches {
				if i > 0 {
					leftPane.WriteString(theme.HelpDescStyle.Render(", "))
				}
				leftPane.WriteString(theme.HelpKeyStyle.Render(p))
			}
			leftPane.WriteString("\n")
		} else if len(matches) > 5 {
			leftPane.WriteString(theme.HelpDescStyle.Render(fmt.Sprintf("  %d matches (type more to filter)\n", len(matches))))
		}
	}

	// Show warnings
	for _, w := range m.warnings {
		leftPane.WriteString(theme.HealthYellowStyle.Render("  ⚠ "+w) + "\n")
	}

	if m.err != "" {
		leftPane.WriteString("\n" + theme.ErrorStyle.Render(m.err))
	}
	leftPane.WriteString("\n" + theme.HelpDescStyle.Render("tab: complete • enter: create • esc: cancel"))

	// Right pane: live JSON preview (pretty-printed)
	var rightPane strings.Builder
	rightPane.WriteString(theme.HelpKeyStyle.Render("Preview:") + "\n\n")
	jsonStr, _ := m.buildJSON()
	if jsonStr == "" {
		rightPane.WriteString(theme.HelpDescStyle.Render("(fill in required fields)"))
	} else {
		var pretty bytes.Buffer
		if json.Indent(&pretty, []byte(jsonStr), "", "  ") == nil {
			rightPane.WriteString(theme.HelpDescStyle.Render(pretty.String()))
		} else {
			rightPane.WriteString(theme.HelpDescStyle.Render(jsonStr))
		}
	}

	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.NewStyle().Width(44).Render(leftPane.String()),
		lipgloss.NewStyle().Width(50).PaddingLeft(2).Render(rightPane.String()),
	)

	b.WriteString(content)

	return theme.ModalStyle.Render(b.String())
}

func (m *Model) checkWarnings() {
	m.warnings = nil
	name := strings.TrimSpace(m.inputs[fieldName].Value())
	patterns := strings.TrimSpace(m.inputs[fieldPatterns].Value())

	if name != "" && !m.editing {
		for _, e := range m.existing {
			if strings.EqualFold(e.Name, name) {
				m.warnings = append(m.warnings, fmt.Sprintf("Template '%s' already exists", name))
				break
			}
		}
	}

	if patterns != "" {
		inputPats := splitPatterns(patterns)
		for _, e := range m.existing {
			// Skip self when editing
			if m.editing && strings.EqualFold(e.Name, strings.TrimSpace(m.inputs[fieldName].Value())) {
				continue
			}
			for _, ip := range inputPats {
				for _, ep := range e.Patterns {
					if patternsOverlap(ip, ep) {
						m.warnings = append(m.warnings, fmt.Sprintf("Pattern '%s' overlaps with template '%s' (%s)", ip, e.Name, ep))
					}
				}
			}
		}
	}
}

func splitPatterns(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// patternsOverlap checks if two glob patterns could match the same index.
// Simple heuristic: strip trailing * and check if one is a prefix of the other.
func patternsOverlap(a, b string) bool {
	a = strings.TrimSuffix(strings.TrimSpace(a), "*")
	b = strings.TrimSuffix(strings.TrimSpace(b), "*")
	return strings.HasPrefix(a, b) || strings.HasPrefix(b, a)
}

func (m Model) matchingPolicies() []string {
	query := strings.ToLower(strings.TrimSpace(m.inputs[fieldILMPolicy].Value()))
	if query == "" {
		return m.ilmPolicies
	}
	var matches []string
	for _, p := range m.ilmPolicies {
		if strings.Contains(strings.ToLower(p), query) {
			matches = append(matches, p)
		}
	}
	return matches
}
