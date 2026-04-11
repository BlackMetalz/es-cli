package shard

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
type ShardsLoadedMsg struct {
	Shards []es.Shard
}

type ErrorMsg struct {
	Err error
}

type SortField int

const (
	SortByIndex SortField = iota
	SortByShard
	SortByState
	SortByNode
	SortByDocs
	SortByStore
)

type Model struct {
	table     table.Model
	shards    []es.Shard
	filtered  []es.Shard
	client    *es.Client
	keys      KeyMap
	sortField SortField
	sortAsc   bool
	showAll   bool
	width     int
	height    int
	colWidths [8]int
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
			{Title: "index", Width: 30},
			{Title: "shard", Width: 8},
			{Title: "prirep", Width: 8},
			{Title: "state", Width: 14},
			{Title: "docs", Width: 12},
			{Title: "store", Width: 12},
			{Title: "ip", Width: 16},
			{Title: "node", Width: 20},
		}),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = theme.TableHeaderStyle
	s.Selected = lipgloss.NewStyle()
	t.SetStyles(s)

	ti := textinput.New()
	ti.Placeholder = "index or node name..."
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
	return m.fetchShards()
}

func (m *Model) fetchShards() tea.Cmd {
	return func() tea.Msg {
		shards, err := m.client.ListShards()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return ShardsLoadedMsg{Shards: shards}
	}
}

func (m *Model) Update(msg tea.Msg) (views.View, tea.Cmd) {
	switch msg := msg.(type) {
	case ShardsLoadedMsg:
		m.shards = msg.Shards
		m.loading = false
		m.err = nil
		m.sortShards()
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

	case key.Matches(msg, m.keys.SortByIndex):
		m.toggleSort(SortByIndex)
		return m, nil

	case key.Matches(msg, m.keys.SortByShard):
		m.toggleSort(SortByShard)
		return m, nil

	case key.Matches(msg, m.keys.SortByState):
		m.toggleSort(SortByState)
		return m, nil

	case key.Matches(msg, m.keys.SortByNode):
		m.toggleSort(SortByNode)
		return m, nil

	case key.Matches(msg, m.keys.SortByDocs):
		m.toggleSort(SortByDocs)
		return m, nil

	case key.Matches(msg, m.keys.SortByStore):
		m.toggleSort(SortByStore)
		return m, nil

	case key.Matches(msg, m.keys.ToggleAll):
		m.showAll = !m.showAll
		m.applyFilter()
		m.updateColumnWidths()
		return m, nil

	case key.Matches(msg, m.keys.Refresh):
		m.loading = true
		return m, m.fetchShards()

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
	m.sortShards()
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
		return "\n  Loading shards..."
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
	switch m.sortField {
	case SortByIndex:
		sortIndicator += theme.HelpKeyStyle.Render("index") + theme.HelpDescStyle.Render(arrow)
	case SortByShard:
		sortIndicator += theme.HelpKeyStyle.Render("shard") + theme.HelpDescStyle.Render(arrow)
	case SortByState:
		sortIndicator += theme.HelpKeyStyle.Render("state") + theme.HelpDescStyle.Render(arrow)
	case SortByNode:
		sortIndicator += theme.HelpKeyStyle.Render("node") + theme.HelpDescStyle.Render(arrow)
	case SortByDocs:
		sortIndicator += theme.HelpKeyStyle.Render("docs") + theme.HelpDescStyle.Render(arrow)
	case SortByStore:
		sortIndicator += theme.HelpKeyStyle.Render("store") + theme.HelpDescStyle.Render(arrow)
	}
	b.WriteString(sortIndicator)

	if m.filter != "" && !m.searching {
		b.WriteString(theme.HelpDescStyle.Render(fmt.Sprintf("  (%d/%d shards)", len(m.filtered), len(m.shards))))
		b.WriteString("  ")
		b.WriteString(theme.HelpKeyStyle.Render("filter: "))
		b.WriteString(theme.HelpDescStyle.Render(m.filter))
	} else {
		b.WriteString(theme.HelpDescStyle.Render(fmt.Sprintf("  (%d shards)", len(m.filtered))))
		if !m.showAll {
			if hidden := m.countHidden(); hidden > 0 {
				b.WriteString(theme.HelpDescStyle.Render(fmt.Sprintf("  %d hidden", hidden)))
			}
		}
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
	return "Shards"
}

func (m *Model) HelpGroups() []views.HelpGroup {
	return []views.HelpGroup{
		{
			Title: "Sort",
			Bindings: []key.Binding{
				m.keys.SortByIndex,
				m.keys.SortByShard,
				m.keys.SortByState,
				m.keys.SortByNode,
				m.keys.SortByDocs,
				m.keys.SortByStore,
			},
		},
		{
			Title: "General",
			Bindings: []key.Binding{
				m.keys.Search,
				m.keys.Refresh,
				m.keys.ToggleAll,
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
	var parts []string
	if m.showAll {
		parts = append(parts, "showing all shards")
	} else {
		parts = append(parts, "hiding system shards")
	}
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
