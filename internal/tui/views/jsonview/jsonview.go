package jsonview

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/kienlt/es-cli/internal/tui/theme"
	"github.com/kienlt/es-cli/internal/tui/views"
)

type DataLoadedMsg struct {
	Data []byte
}

type ErrorMsg struct {
	Err error
}

// GoBackMsg signals that the user wants to go back.
type GoBackMsg struct{}

type KeyMap struct {
	Back key.Binding
	Quit key.Binding
}

type Model struct {
	title    string
	viewport viewport.Model
	data     []byte
	width    int
	height   int
	loading  bool
	err      error
	keys     KeyMap
	fetchCmd func() tea.Msg
}

var _ views.View = (*Model)(nil)

// New creates a JSON detail view. fetchCmd is called on Init to fetch data.
func New(title string, fetchCmd func() tea.Msg) *Model {
	return &Model{
		title:    title,
		loading:  true,
		fetchCmd: fetchCmd,
		keys: KeyMap{
			Back: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
			Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		},
	}
}

func (m *Model) Init() tea.Cmd {
	return func() tea.Msg { return m.fetchCmd() }
}

func (m *Model) Update(msg tea.Msg) (views.View, tea.Cmd) {
	switch msg := msg.(type) {
	case DataLoadedMsg:
		m.data = msg.Data
		m.loading = false
		m.err = nil
		m.updateViewport()
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return GoBackMsg{} }
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	if m.loading {
		return "\n  Loading..."
	}
	if m.err != nil {
		return "\n  " + theme.ErrorStyle.Render("Error: "+m.err.Error())
	}
	return m.viewport.View()
}

func (m *Model) Name() string                           { return m.title }
func (m *Model) IsInputMode() bool                      { return false }
func (m *Model) PopPendingAction() *views.PendingAction { return nil }
func (m *Model) StatusInfo() string                     { return m.title }

func (m *Model) HelpGroups() []views.HelpGroup {
	return []views.HelpGroup{
		{
			Title:    "Navigation",
			Bindings: []key.Binding{m.keys.Back, m.keys.Quit},
		},
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.viewport.Width = width
	m.viewport.Height = height
	if m.data != nil {
		m.updateViewport()
	}
}

func (m *Model) updateViewport() {
	var buf bytes.Buffer
	if err := json.Indent(&buf, m.data, "", "  "); err != nil {
		m.viewport.SetContent(" " + string(m.data))
		return
	}
	content := buf.String()
	m.viewport.SetContent(" " + strings.ReplaceAll(content, "\n", "\n "))
	m.viewport.GotoTop()
}
