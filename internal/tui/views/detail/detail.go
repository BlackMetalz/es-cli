package detail

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kienlt/es-cli/internal/es"
	"github.com/kienlt/es-cli/internal/tui/theme"
	"github.com/kienlt/es-cli/internal/tui/views"
)

type Tab int

const (
	TabSettings Tab = iota
	TabMappings
	TabAliases
)

var tabNames = []string{"Settings", "Mappings", "Aliases"}

type DetailLoadedMsg struct {
	Detail *es.IndexDetail
}

type ErrorMsg struct {
	Err error
}

// GoBackMsg signals that the user wants to go back to the index list.
type GoBackMsg struct{}

type KeyMap struct {
	NextTab key.Binding
	PrevTab key.Binding
	Back    key.Binding
	Quit    key.Binding
}

func defaultKeyMap() KeyMap {
	return KeyMap{
		NextTab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev tab"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

type Model struct {
	indexName string
	client    *es.Client
	keys      KeyMap
	viewport  viewport.Model
	tab       Tab
	detail    *es.IndexDetail
	width     int
	height    int
	loading   bool
	err       error
}

var _ views.View = (*Model)(nil)

func New(client *es.Client, indexName string) *Model {
	return &Model{
		indexName: indexName,
		client:    client,
		keys:      defaultKeyMap(),
		loading:   true,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.fetchDetail()
}

func (m *Model) fetchDetail() tea.Cmd {
	name := m.indexName
	return func() tea.Msg {
		detail, err := m.client.GetIndexDetail(name)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return DetailLoadedMsg{Detail: detail}
	}
}

func (m *Model) Update(msg tea.Msg) (views.View, tea.Cmd) {
	switch msg := msg.(type) {
	case DetailLoadedMsg:
		m.detail = msg.Detail
		m.loading = false
		m.err = nil
		m.updateViewport()
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) handleKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Back):
		return m, func() tea.Msg { return GoBackMsg{} }

	case key.Matches(msg, m.keys.NextTab):
		m.tab = (m.tab + 1) % Tab(len(tabNames))
		m.updateViewport()
		return m, nil

	case key.Matches(msg, m.keys.PrevTab):
		m.tab = (m.tab - 1 + Tab(len(tabNames))) % Tab(len(tabNames))
		m.updateViewport()
		return m, nil

	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	if m.loading {
		return "\n  Loading index detail..."
	}
	if m.err != nil {
		return "\n  " + theme.ErrorStyle.Render("Error: "+m.err.Error())
	}

	var b strings.Builder

	// Tab bar
	b.WriteString(" ")
	for i, name := range tabNames {
		if Tab(i) == m.tab {
			b.WriteString(theme.ViewNameStyle.Render(name))
		} else {
			b.WriteString(theme.HelpDescStyle.Render(" " + name + " "))
		}
		b.WriteString(" ")
	}
	b.WriteString("\n\n")

	// Viewport content
	b.WriteString(m.viewport.View())

	return b.String()
}

func (m *Model) Name() string {
	return "Index: " + m.indexName
}

func (m *Model) HelpGroups() []views.HelpGroup {
	return []views.HelpGroup{
		{
			Title: "Navigation",
			Bindings: []key.Binding{
				m.keys.NextTab,
				m.keys.PrevTab,
				m.keys.Back,
				m.keys.Quit,
			},
		},
	}
}

func (m *Model) IsInputMode() bool {
	return false
}

func (m *Model) PopPendingAction() *views.PendingAction {
	return nil
}

func (m *Model) StatusInfo() string {
	return m.indexName
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height - 3 // tab bar + spacing
	if m.detail != nil {
		m.updateViewport()
	}
}

func (m *Model) updateViewport() {
	if m.detail == nil {
		return
	}

	var raw json.RawMessage
	switch m.tab {
	case TabSettings:
		raw = m.detail.Settings
	case TabMappings:
		raw = m.detail.Mappings
	case TabAliases:
		raw = m.detail.Aliases
	}

	content := prettyJSON(raw)
	m.viewport.SetContent(" " + strings.ReplaceAll(content, "\n", "\n "))
	m.viewport.GotoTop()
}

func prettyJSON(data json.RawMessage) string {
	if len(data) == 0 {
		return "{}"
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err != nil {
		return string(data)
	}
	return colorizeJSON(buf.String())
}

// colorizeJSON adds basic syntax coloring to JSON output.
func colorizeJSON(s string) string {
	var b strings.Builder
	inString := false
	isKey := false
	escaped := false

	for i := 0; i < len(s); i++ {
		c := s[i]

		if escaped {
			b.WriteByte(c)
			escaped = false
			continue
		}

		if c == '\\' && inString {
			b.WriteByte(c)
			escaped = true
			continue
		}

		if c == '"' {
			if !inString {
				inString = true
				// Detect if this is a key (next non-space after value is ':')
				isKey = isJSONKey(s, i)
				if isKey {
					b.WriteString(theme.HelpKeyStyle.Render("\""))
				} else {
					b.WriteString(theme.HealthGreenStyle.Render("\""))
				}
			} else {
				if isKey {
					b.WriteString(theme.HelpKeyStyle.Render("\""))
				} else {
					b.WriteString(theme.HealthGreenStyle.Render("\""))
				}
				inString = false
			}
			continue
		}

		if inString {
			if isKey {
				b.WriteString(theme.HelpKeyStyle.Render(string(c)))
			} else {
				b.WriteString(theme.HealthGreenStyle.Render(string(c)))
			}
			continue
		}

		// Numbers
		if c >= '0' && c <= '9' {
			b.WriteString(theme.HealthYellowStyle.Render(string(c)))
			continue
		}

		// Booleans and null
		for _, keyword := range []string{"true", "false", "null"} {
			if strings.HasPrefix(s[i:], keyword) {
				b.WriteString(theme.HealthYellowStyle.Render(keyword))
				i += len(keyword) - 1
				c = 0
				break
			}
		}
		if c == 0 {
			continue
		}

		b.WriteByte(c)
	}
	return b.String()
}

func isJSONKey(s string, quotePos int) bool {
	// Find the closing quote
	for i := quotePos + 1; i < len(s); i++ {
		if s[i] == '\\' {
			i++
			continue
		}
		if s[i] == '"' {
			// Check if followed by ':'
			for j := i + 1; j < len(s); j++ {
				if s[j] == ' ' || s[j] == '\t' {
					continue
				}
				return s[j] == ':'
			}
			return false
		}
	}
	return false
}
