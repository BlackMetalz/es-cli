package task

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

type TasksLoadedMsg struct {
	Tasks []es.Task
	Total int
}

type ErrorMsg struct {
	Err error
}

type ActionCompleteMsg struct {
	TaskID string
}

type Model struct {
	table         table.Model
	tasks         []es.Task
	filtered      []es.Task
	filteredCount int
	total         int

	client    *es.Client
	keys      KeyMap
	sortField SortField
	sortAsc   bool
	showAll   bool
	category  CategoryFilter

	width     int
	height    int
	colWidths [6]int
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
			{Title: "duration", Width: defaultColWidths[colDuration]},
			{Title: "node", Width: defaultColWidths[colNode]},
			{Title: "action", Width: defaultColWidths[colAction]},
			{Title: "type", Width: defaultColWidths[colType]},
			{Title: "cancelable", Width: defaultColWidths[colCancellable]},
			{Title: "description", Width: defaultColWidths[colDescription]},
		}),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = theme.TableHeaderStyle
	// Reverse on the cursor row is a sentinel for postProcessTable to locate it
	// independent of viewport offset and duplicate row content. postProcessTable
	// strips the reverse wrap before output, replacing it with our own highlight.
	s.Selected = lipgloss.NewStyle().Reverse(true)
	t.SetStyles(s)

	ti := textinput.New()
	ti.Placeholder = "action, node, or description..."
	ti.Prompt = "/ "
	ti.CharLimit = 256

	return &Model{
		table:       t,
		client:      client,
		keys:        keys,
		loading:     true,
		sortField:   SortByDuration,
		sortAsc:     false,
		searchInput: ti,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.fetchTasks()
}

func (m *Model) fetchTasks() tea.Cmd {
	return func() tea.Msg {
		tasks, total, err := m.client.ListTasks()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return TasksLoadedMsg{Tasks: tasks, Total: total}
	}
}

func (m *Model) selectedTask() *es.Task {
	cursor := m.table.Cursor()
	if cursor < 0 || cursor >= len(m.filtered) {
		return nil
	}
	return &m.filtered[cursor]
}

func (m *Model) Update(msg tea.Msg) (views.View, tea.Cmd) {
	switch msg := msg.(type) {
	case TasksLoadedMsg:
		m.tasks = msg.Tasks
		m.total = msg.Total
		m.loading = false
		m.err = nil
		m.sortTasks()
		m.applyFilter()
		m.updateColumnWidths()
		return m, nil

	case ActionCompleteMsg:
		m.loading = true
		return m, m.fetchTasks()

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
	case key.Matches(msg, m.keys.Up), key.Matches(msg, m.keys.Down):
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		return m, cmd
	case key.Matches(msg, m.keys.SortByDuration):
		m.toggleSort(SortByDuration)
	case key.Matches(msg, m.keys.SortByAction):
		m.toggleSort(SortByAction)
	case key.Matches(msg, m.keys.SortByNode):
		m.toggleSort(SortByNode)
	case key.Matches(msg, m.keys.CycleCategory):
		m.nextCategory()
	case key.Matches(msg, m.keys.ToggleAll):
		m.showAll = !m.showAll
		m.applyFilter()
		m.updateColumnWidths()
	case key.Matches(msg, m.keys.Cancel):
		if sel := m.selectedTask(); sel != nil && sel.Cancellable {
			m.pendingAction = &views.PendingAction{Type: "cancel_task", Index: sel.ID}
		}
	case key.Matches(msg, m.keys.Detail):
		if sel := m.selectedTask(); sel != nil {
			m.pendingAction = &views.PendingAction{Type: "view_task_detail", Index: sel.ID}
		}
	case key.Matches(msg, m.keys.Refresh):
		m.loading = true
		return m, m.fetchTasks()
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
		return "\n  Loading tasks..."
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
		SortByDuration: "duration",
		SortByAction:   "action",
		SortByNode:     "node",
	}
	b.WriteString(" Sort: ")
	b.WriteString(theme.HelpKeyStyle.Render(sortLabel[m.sortField]))
	b.WriteString(theme.HelpDescStyle.Render(arrow))

	if m.category != CategoryAll {
		b.WriteString("  ")
		b.WriteString(theme.HelpKeyStyle.Render("category: "))
		b.WriteString(theme.HelpDescStyle.Render(m.category.String()))
	}

	b.WriteString("  ")
	if m.showAll {
		b.WriteString(theme.HealthGreenStyle.Render("[all]"))
	} else {
		b.WriteString(theme.HelpDescStyle.Render(fmt.Sprintf("[top %d]", topN)))
	}

	if m.filter != "" && !m.searching {
		b.WriteString(theme.HelpDescStyle.Render(fmt.Sprintf("  filter: %s", m.filter)))
	}

	if !m.showAll && m.filteredCount > topN {
		b.WriteString(theme.HelpDescStyle.Render(
			fmt.Sprintf("  %d/%d tasks (%d total)", topN, m.filteredCount, m.total)))
	} else {
		b.WriteString(theme.HelpDescStyle.Render(
			fmt.Sprintf("  %d tasks (%d total)", m.filteredCount, m.total)))
	}

	b.WriteString("\n")

	if m.searching {
		b.WriteString(" " + m.searchInput.View())
	}
	b.WriteString("\n")

	b.WriteString(m.postProcessTable(m.table.View()))
	return b.String()
}

func (m *Model) Name() string { return "Tasks" }

func (m *Model) HelpGroups() []views.HelpGroup {
	return []views.HelpGroup{
		{
			Title: "Navigate",
			Bindings: []key.Binding{
				m.keys.Up,
				m.keys.Down,
			},
		},
		{
			Title: "Sort",
			Bindings: []key.Binding{
				m.keys.SortByDuration,
				m.keys.SortByAction,
				m.keys.SortByNode,
			},
		},
		{
			Title: "Actions",
			Bindings: []key.Binding{
				m.keys.Detail,
				m.keys.Cancel,
			},
		},
		{
			Title: "General",
			Bindings: []key.Binding{
				m.keys.Search,
				m.keys.CycleCategory,
				m.keys.ToggleAll,
				m.keys.Refresh,
				m.keys.Help,
				m.keys.Quit,
			},
		},
	}
}

func (m *Model) IsInputMode() bool { return m.searching }

func (m *Model) PopPendingAction() *views.PendingAction {
	a := m.pendingAction
	m.pendingAction = nil
	return a
}

func (m *Model) StatusInfo() string {
	if m.total == 0 {
		return ""
	}
	parts := []string{}
	if !m.showAll && m.filteredCount > topN {
		parts = append(parts, fmt.Sprintf("Top %d of %d tasks (%d total)", topN, m.filteredCount, m.total))
	} else {
		parts = append(parts, fmt.Sprintf("%d tasks (%d total)", m.filteredCount, m.total))
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
