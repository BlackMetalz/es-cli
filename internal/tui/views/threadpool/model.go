package threadpool

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

type ThreadPoolsLoadedMsg struct {
	Pools []es.ThreadPool
}

type ErrorMsg struct {
	Err error
}

type Model struct {
	table     table.Model
	pools     []es.ThreadPool
	filtered  []es.ThreadPool
	client    *es.Client
	keys      KeyMap
	sortField SortField
	sortAsc   bool
	showAll   bool // include "direct" type pools when true
	width     int
	height    int
	colWidths [8]int
	err       error
	loading   bool

	searching   bool
	searchInput textinput.Model
	filter      string

	pendingAction *views.PendingAction
}

var _ views.View = (*Model)(nil)

func New(client *es.Client) *Model {
	keys := DefaultKeyMap()

	t := table.New(
		table.WithColumns([]table.Column{
			{Title: "node", Width: defaultColWidths[colNode]},
			{Title: "name", Width: defaultColWidths[colName]},
			{Title: "type", Width: defaultColWidths[colType]},
			{Title: "active", Width: defaultColWidths[colActive]},
			{Title: "size", Width: defaultColWidths[colSize]},
			{Title: "queue", Width: defaultColWidths[colQueue]},
			{Title: "rejected", Width: defaultColWidths[colRejected]},
			{Title: "largest", Width: defaultColWidths[colLargest]},
		}),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = theme.TableHeaderStyle
	s.Selected = lipgloss.NewStyle()
	t.SetStyles(s)

	ti := textinput.New()
	ti.Placeholder = "pool or node name..."
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
	return m.fetchThreadPools()
}

func (m *Model) fetchThreadPools() tea.Cmd {
	return func() tea.Msg {
		pools, err := m.client.ListThreadPools()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return ThreadPoolsLoadedMsg{Pools: pools}
	}
}

func (m *Model) Update(msg tea.Msg) (views.View, tea.Cmd) {
	switch msg := msg.(type) {
	case ThreadPoolsLoadedMsg:
		m.pools = msg.Pools
		m.loading = false
		m.err = nil
		m.sortPools()
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
	case key.Matches(msg, m.keys.SortByName):
		m.toggleSort(SortByName)
	case key.Matches(msg, m.keys.SortByNode):
		m.toggleSort(SortByNode)
	case key.Matches(msg, m.keys.SortByActive):
		m.toggleSort(SortByActive)
	case key.Matches(msg, m.keys.SortByQueue):
		m.toggleSort(SortByQueue)
	case key.Matches(msg, m.keys.SortByRejected):
		m.toggleSort(SortByRejected)
	case key.Matches(msg, m.keys.ToggleAll):
		m.showAll = !m.showAll
		m.applyFilter()
		m.updateColumnWidths()
	case key.Matches(msg, m.keys.Refresh):
		m.loading = true
		return m, m.fetchThreadPools()
	case key.Matches(msg, m.keys.Search):
		m.searching = true
		m.searchInput.SetValue(m.filter)
		m.searchInput.Focus()
		return m, m.searchInput.Cursor.BlinkCmd()
	case key.Matches(msg, m.keys.Help):
		return m, nil
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	}
	return m, nil
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

func (m *Model) View() string {
	if m.loading {
		return "\n  Loading thread pools..."
	}
	if m.err != nil {
		return "\n  " + theme.ErrorStyle.Render("Error: "+m.err.Error())
	}

	var b strings.Builder

	arrow := " ↑"
	if !m.sortAsc {
		arrow = " ↓"
	}
	sortLabel := map[SortField]string{
		SortByName:     "name",
		SortByNode:     "node",
		SortByActive:   "active",
		SortByQueue:    "queue",
		SortByRejected: "rejected",
	}
	b.WriteString(" Sort: ")
	b.WriteString(theme.HelpKeyStyle.Render(sortLabel[m.sortField]))
	b.WriteString(theme.HelpDescStyle.Render(arrow))

	if m.filter != "" && !m.searching {
		b.WriteString(theme.HelpDescStyle.Render(fmt.Sprintf("  (%d/%d pools)", len(m.filtered), len(m.pools))))
		b.WriteString("  ")
		b.WriteString(theme.HelpKeyStyle.Render("filter: "))
		b.WriteString(theme.HelpDescStyle.Render(m.filter))
	} else {
		b.WriteString(theme.HelpDescStyle.Render(fmt.Sprintf("  (%d pools)", len(m.filtered))))
		if !m.showAll {
			if hidden := m.countHidden(); hidden > 0 {
				b.WriteString(theme.HelpDescStyle.Render(fmt.Sprintf("  %d direct hidden", hidden)))
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

func (m *Model) Name() string { return "Thread Pools" }

func (m *Model) HelpGroups() []views.HelpGroup {
	return []views.HelpGroup{
		{
			Title: "Sort",
			Bindings: []key.Binding{
				m.keys.SortByName,
				m.keys.SortByNode,
				m.keys.SortByActive,
				m.keys.SortByQueue,
				m.keys.SortByRejected,
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
	var issues int
	for _, tp := range m.filtered {
		if tp.Rejected > 0 || tp.Queue > 0 {
			issues++
		}
	}
	var parts []string
	if issues > 0 {
		parts = append(parts, fmt.Sprintf("%d pool(s) with issues", issues))
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
