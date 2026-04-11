package node

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kienlt/es-cli/internal/es"
	"github.com/kienlt/es-cli/internal/tui/theme"
	"github.com/kienlt/es-cli/internal/tui/views"
)

// Messages
type NodesLoadedMsg struct {
	Nodes []es.Node
}

type ErrorMsg struct {
	Err error
}

type Model struct {
	table     table.Model
	nodes     []es.Node
	filtered  []es.Node
	client    *es.Client
	keys      KeyMap
	sortField SortField
	sortAsc   bool
	width     int
	height    int
	colWidths [11]int
	err       error
	loading   bool

	// Search
	searching   bool
	searchInput textinput.Model
	filter      string

	// Pending action for app to handle
	pendingAction *views.PendingAction
}

var _ views.View = (*Model)(nil)

func New(client *es.Client) *Model {
	keys := DefaultKeyMap()

	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "name", Width: 20},
			{Title: "ip", Width: 16},
			{Title: "heap%", Width: 8},
			{Title: "ram%", Width: 8},
			{Title: "cpu", Width: 6},
			{Title: "load_1m", Width: 8},
			{Title: "load_5m", Width: 8},
			{Title: "load_15m", Width: 8},
			{Title: "role", Width: 14},
			{Title: "master", Width: 8},
			{Title: "disk%", Width: 8},
		}),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = theme.TableHeaderStyle
	s.Selected = lipgloss.NewStyle()
	t.SetStyles(s)

	ti := textinput.New()
	ti.Placeholder = "node name..."
	ti.Prompt = "/ "
	ti.CharLimit = 256

	return &Model{
		table:       t,
		client:      client,
		keys:        keys,
		loading:     true,
		sortAsc:     true,
		searchInput: ti,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.fetchNodes()
}

func (m *Model) fetchNodes() tea.Cmd {
	return func() tea.Msg {
		nodes, err := m.client.ListNodes()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return NodesLoadedMsg{Nodes: nodes}
	}
}

func (m *Model) Update(msg tea.Msg) (views.View, tea.Cmd) {
	switch msg := msg.(type) {
	case NodesLoadedMsg:
		m.nodes = msg.Nodes
		m.loading = false
		m.err = nil
		m.sortNodes()
		m.applyFilter()
		m.updateColumnWidths()
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// When searching, route non-key messages (like blink) to the text input
	if m.searching {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *Model) handleKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	if m.searching {
		return m.handleSearchKey(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Search):
		m.searching = true
		m.searchInput.SetValue(m.filter)
		m.searchInput.Focus()
		return m, m.searchInput.Cursor.BlinkCmd()

	case key.Matches(msg, m.keys.Help):
		return m, nil

	case key.Matches(msg, m.keys.SortByName):
		m.toggleSort(SortByName)
		return m, nil

	case key.Matches(msg, m.keys.SortByCPU):
		m.toggleSort(SortByCPU)
		return m, nil

	case key.Matches(msg, m.keys.SortByHeap):
		m.toggleSort(SortByHeap)
		return m, nil

	case key.Matches(msg, m.keys.SortByRAM):
		m.toggleSort(SortByRAM)
		return m, nil

	case key.Matches(msg, m.keys.SortByDisk):
		m.toggleSort(SortByDisk)
		return m, nil

	case key.Matches(msg, m.keys.Maintenance):
		m.pendingAction = &views.PendingAction{Type: "set_allocation"}
		return m, nil

	case key.Matches(msg, m.keys.Refresh):
		m.loading = true
		return m, m.fetchNodes()

	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *Model) toggleSort(field SortField) {
	if m.sortField == field {
		m.sortAsc = !m.sortAsc
	} else {
		m.sortField = field
		m.sortAsc = true
	}
	m.sortNodes()
	m.applyFilter()
	m.updateTable()
}

func (m *Model) handleSearchKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		m.filter = m.searchInput.Value()
		m.searching = false
		m.searchInput.Blur()
		m.applyFilter()
		m.updateTable()
		return m, nil

	case tea.KeyEsc:
		m.searching = false
		m.filter = ""
		m.searchInput.SetValue("")
		m.searchInput.Blur()
		m.applyFilter()
		m.updateTable()
		return m, nil
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	m.filter = m.searchInput.Value()
	m.applyFilter()
	m.updateTable()
	return m, cmd
}

// --- View interface ---

func (m *Model) View() string {
	if m.loading {
		return "\n  Loading nodes..."
	}
	if m.err != nil {
		return "\n  " + theme.ErrorStyle.Render("Error: "+m.err.Error())
	}

	var b strings.Builder

	// Sort indicator
	arrow := " ↑"
	if !m.sortAsc {
		arrow = " ↓"
	}
	sortIndicator := " Sort: "
	sortIndicator += theme.HelpKeyStyle.Render(sortFieldLabel(m.sortField))
	sortIndicator += theme.HelpDescStyle.Render(arrow)
	b.WriteString(sortIndicator)

	if m.filter != "" && !m.searching {
		b.WriteString(theme.HelpDescStyle.Render(fmt.Sprintf("  (%d/%d nodes)", len(m.filtered), len(m.nodes))))
		b.WriteString("  ")
		b.WriteString(theme.HelpKeyStyle.Render("filter: "))
		b.WriteString(theme.HelpDescStyle.Render(m.filter))
	} else {
		b.WriteString(theme.HelpDescStyle.Render(fmt.Sprintf("  (%d nodes)", len(m.filtered))))
	}
	b.WriteString("\n")

	if m.searching {
		b.WriteString(" " + m.searchInput.View())
	}
	b.WriteString("\n")

	b.WriteString(m.postProcessTable(m.table.View()))

	return b.String()
}

func (m *Model) Name() string {
	return "Nodes"
}

func (m *Model) HelpGroups() []views.HelpGroup {
	return []views.HelpGroup{
		{
			Title: "Sort",
			Bindings: []key.Binding{
				m.keys.SortByName,
				m.keys.SortByCPU,
				m.keys.SortByHeap,
				m.keys.SortByRAM,
				m.keys.SortByDisk,
			},
		},
		{
			Title: "Node",
			Bindings: []key.Binding{
				m.keys.Maintenance,
				m.keys.Refresh,
			},
		},
		{
			Title: "General",
			Bindings: []key.Binding{
				m.keys.Search,
				m.keys.Help,
				m.keys.Quit,
			},
		},
	}
}

func (m *Model) IsInputMode() bool {
	return m.searching
}

func (m *Model) PopPendingAction() *views.PendingAction {
	a := m.pendingAction
	m.pendingAction = nil
	return a
}

func (m *Model) StatusInfo() string {
	parts := []string{fmt.Sprintf("%d nodes", len(m.filtered))}
	if m.filter != "" {
		parts = append(parts, fmt.Sprintf("filter: %s", m.filter))
	}
	return strings.Join(parts, " | ")
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.table.SetHeight(height - 5)
	m.updateColumnWidths()
}
