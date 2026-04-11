package index

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
type IndicesLoadedMsg struct {
	Indices []es.Index
}

type ErrorMsg struct {
	Err error
}

type ActionCompleteMsg struct {
	Action string
	Index  string
}

type SortField int

const (
	SortByName SortField = iota
	SortByHealth
	SortBySize
	SortByDocs
)

type Model struct {
	table     table.Model
	indices   []es.Index
	filtered  []es.Index
	client    *es.Client
	keys      KeyMap
	sortField SortField
	sortAsc   bool
	showAll   bool
	width     int
	height    int
	colWidths [7]int
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
			{Title: "health", Width: 8},
			{Title: "status", Width: 8},
			{Title: "index", Width: 40},
			{Title: "primary shard", Width: 15},
			{Title: "replica shard", Width: 15},
			{Title: "docs.count", Width: 12},
			{Title: "pri.store.size", Width: 16},
		}),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = theme.TableHeaderStyle
	// Use empty selected style — we post-process for full-width highlight
	s.Selected = lipgloss.NewStyle()
	t.SetStyles(s)

	ti := textinput.New()
	ti.Placeholder = "index name..."
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
	return m.fetchIndices()
}

func (m *Model) fetchIndices() tea.Cmd {
	return func() tea.Msg {
		indices, err := m.client.ListIndices()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return IndicesLoadedMsg{Indices: indices}
	}
}

func (m *Model) Update(msg tea.Msg) (views.View, tea.Cmd) {
	switch msg := msg.(type) {
	case IndicesLoadedMsg:
		m.indices = msg.Indices
		m.loading = false
		m.err = nil
		m.sortIndices()
		m.applyFilter()
		m.updateColumnWidths()
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		m.loading = false
		return m, nil

	case ActionCompleteMsg:
		m.pendingAction = nil
		m.filter = ""
		m.searchInput.SetValue("")
		m.loading = true
		return m, m.fetchIndices()

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
		// Handled by app.go (full-screen help overlay)
		return m, nil

	case key.Matches(msg, m.keys.ToggleAll):
		m.showAll = !m.showAll
		m.applyFilter()
		m.updateColumnWidths()
		return m, nil

	case key.Matches(msg, m.keys.ViewDetail):
		if selected := m.selectedIndex(); selected != "" {
			m.pendingAction = &views.PendingAction{Type: "view_detail", Index: selected}
		}
		return m, nil

	case key.Matches(msg, m.keys.SortByHealth):
		m.toggleSort(SortByHealth)
		return m, nil

	case key.Matches(msg, m.keys.SortByName):
		m.toggleSort(SortByName)
		return m, nil

	case key.Matches(msg, m.keys.SortBySize):
		m.toggleSort(SortBySize)
		return m, nil

	case key.Matches(msg, m.keys.SortByDocs):
		m.toggleSort(SortByDocs)
		return m, nil

	case key.Matches(msg, m.keys.ToggleOpenClose):
		if selected := m.selectedIndex(); selected != "" {
			action := "close"
			if m.selectedStatus() == "close" {
				action = "open"
			}
			m.pendingAction = &views.PendingAction{Type: action, Index: selected}
		}
		return m, nil

	case key.Matches(msg, m.keys.DeleteIndex):
		if selected := m.selectedIndex(); selected != "" {
			m.pendingAction = &views.PendingAction{Type: "delete", Index: selected}
		}
		return m, nil

	case key.Matches(msg, m.keys.CreateIndex):
		return m, nil

	case key.Matches(msg, m.keys.Refresh):
		m.loading = true
		return m, m.fetchIndices()

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
	m.sortIndices()
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
		return "\n  Loading indices..."
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
	case SortByName:
		sortIndicator += theme.HelpKeyStyle.Render("index") + theme.HelpDescStyle.Render(arrow)
	case SortByHealth:
		sortIndicator += theme.HelpKeyStyle.Render("health") + theme.HelpDescStyle.Render(arrow)
	case SortBySize:
		sortIndicator += theme.HelpKeyStyle.Render("pri.store.size") + theme.HelpDescStyle.Render(arrow)
	case SortByDocs:
		sortIndicator += theme.HelpKeyStyle.Render("docs.count") + theme.HelpDescStyle.Render(arrow)
	}
	b.WriteString(sortIndicator)

	if m.filter != "" && !m.searching {
		b.WriteString(theme.HelpDescStyle.Render(fmt.Sprintf("  (%d/%d indices)", len(m.filtered), len(m.indices))))
		b.WriteString("  ")
		b.WriteString(theme.HelpKeyStyle.Render("filter: "))
		b.WriteString(theme.HelpDescStyle.Render(m.filter))
	} else {
		b.WriteString(theme.HelpDescStyle.Render(fmt.Sprintf("  (%d indices)", len(m.filtered))))
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
	return "Indices"
}

func (m *Model) HelpGroups() []views.HelpGroup {
	return []views.HelpGroup{
		{
			Title: "Sort",
			Bindings: []key.Binding{
				m.keys.SortByName,
				m.keys.SortByHealth,
				m.keys.SortBySize,
				m.keys.SortByDocs,
			},
		},
		{
			Title: "Index",
			Bindings: []key.Binding{
				m.keys.ViewDetail,
				m.keys.CreateIndex,
				m.keys.ToggleOpenClose,
				m.keys.DeleteIndex,
				m.keys.Refresh,
				m.keys.ToggleAll,
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
	var parts []string
	if m.showAll {
		parts = append(parts, "showing all indices")
	} else {
		parts = append(parts, "hiding system indices")
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

func (m *Model) selectedIndex() string {
	row := m.table.SelectedRow()
	if row == nil || len(row) < 3 {
		return ""
	}
	return row[2]
}

func (m *Model) selectedStatus() string {
	row := m.table.SelectedRow()
	if row == nil || len(row) < 2 {
		return ""
	}
	return row[1]
}
