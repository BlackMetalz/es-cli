package query

import (
	"fmt"
	"strings"
	"time"

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
type IndicesLoadedMsg struct{ Names []string }
type MappingLoadedMsg struct{ Fields []es.FieldMapping }
type SearchResultMsg struct{ Result *es.SearchResult }
type ErrorMsg struct{ Err error }
type FollowTickMsg struct{}

const pageSize = 500

var followIntervals = []time.Duration{0, 1 * time.Second, 2 * time.Second, 5 * time.Second, 10 * time.Second}

type timeRange struct {
	Label    string
	Duration time.Duration // 0 = all time
}

var timeRanges = []timeRange{
	{"Last 15m", 15 * time.Minute},
	{"Last 1h", time.Hour},
	{"Last 24h", 24 * time.Hour},
	{"Last 7d", 7 * 24 * time.Hour},
	{"Last 30d", 30 * 24 * time.Hour},
	{"All time", 0},
	{"Relative", -1}, // sentinel: relative duration input (e.g. 10d, 2h)
	{"Absolute", -2}, // sentinel: absolute from/to datetime input
}

type Model struct {
	client *es.Client
	keys   KeyMap

	// Index selection
	selectingIndex bool
	indices        []string
	filteredIdx    []string
	indexCursor    int
	indexSearching bool
	indexInput     textinput.Model

	// Query state
	index      string
	query      string
	editing    bool
	queryInput textinput.Model
	columns    []string
	allFields  []es.FieldMapping

	// Results + pagination
	table     table.Model
	hits      []es.SearchHit
	total     int64
	page      int             // current page (0-based)
	pageSorts [][]interface{} // search_after value for each page boundary

	// Follow
	followInterval time.Duration
	followIdx      int

	// Time range picker
	timeRangeIdx     int // current active time range
	pickingTime      bool
	timeCursor       int
	timeCustomInput  textinput.Model
	timeCustomActive bool

	// Custom range (absolute from/to)
	timeRangeActive bool // true when entering from/to
	timeRangeField  int  // 0 = from, 1 = to
	timeFromInput   textinput.Model
	timeToInput     textinput.Model
	timeFrom        time.Time // stored absolute from
	timeTo          time.Time // stored absolute to

	// Column picker
	pickingColumns bool
	columnCursor   int
	columnSelected map[string]bool

	// Query builder
	buildingQuery    bool
	qbSt             qbState
	qbFilters        []FilterCondition
	qbCursor         int
	qbFieldCursor    int
	qbFilteredFields []es.FieldMapping
	qbFieldInput     textinput.Model
	qbValueInput     textinput.Model
	qbPickedField    string

	// Doc detail overlay
	viewingDoc bool
	docContent string
	docScroll  int

	// Standard
	width, height int
	loading       bool
	err           error
	pendingAction *views.PendingAction
}

var _ views.View = (*Model)(nil)

func New(client *es.Client) *Model {
	keys := DefaultKeyMap()

	t := table.New(table.WithFocused(true))
	s := table.DefaultStyles()
	s.Header = theme.TableHeaderStyle
	s.Selected = lipgloss.NewStyle()
	t.SetStyles(s)

	qi := textinput.New()
	qi.Placeholder = "e.g. level:ERROR AND service:api  |  message:\"timeout\" OR status_code:500"
	qi.Prompt = "raw> "
	qi.CharLimit = 512

	ii := textinput.New()
	ii.Placeholder = "filter indices..."
	ii.Prompt = "/ "
	ii.CharLimit = 256

	fi := textinput.New()
	fi.Placeholder = "type to filter fields..."
	fi.Prompt = "> "
	fi.CharLimit = 128

	vi := textinput.New()
	vi.Placeholder = "value..."
	vi.Prompt = "= "
	vi.CharLimit = 256

	ti := textinput.New()
	ti.Placeholder = "e.g. 10d, 2h, 30m"
	ti.Prompt = "> "
	ti.CharLimit = 20

	fromIn := textinput.New()
	fromIn.Placeholder = "2026-03-30 13:00:00"
	fromIn.Prompt = "From> "
	fromIn.CharLimit = 30

	toIn := textinput.New()
	toIn.Placeholder = "2026-03-30 14:30:00"
	toIn.Prompt = "  To> "
	toIn.CharLimit = 30

	return &Model{
		client:          client,
		keys:            keys,
		selectingIndex:  true,
		loading:         true,
		table:           t,
		queryInput:      qi,
		indexInput:      ii,
		qbFieldInput:    fi,
		qbValueInput:    vi,
		timeCustomInput: ti,
		timeFromInput:   fromIn,
		timeToInput:     toIn,
		columnSelected:  make(map[string]bool),
	}
}

func (m *Model) Init() tea.Cmd {
	return m.fetchIndices()
}

func (m *Model) fetchIndices() tea.Cmd {
	return func() tea.Msg {
		names, err := m.client.GetIndexNames()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return IndicesLoadedMsg{Names: names}
	}
}

func (m *Model) fetchMapping() tea.Cmd {
	idx := m.index
	return func() tea.Msg {
		fields, err := m.client.GetFieldMapping(idx)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return MappingLoadedMsg{Fields: fields}
	}
}

func (m *Model) buildFullQuery() string {
	q := m.query
	tr := timeRanges[m.timeRangeIdx]
	var timeFilter string
	if tr.Duration == -2 && !m.timeFrom.IsZero() {
		// Absolute custom range
		from := m.timeFrom.UTC().Format(time.RFC3339)
		to := m.timeTo.UTC().Format(time.RFC3339)
		if m.timeTo.IsZero() {
			to = "*"
		}
		timeFilter = "@timestamp:[" + from + " TO " + to + "]"
	} else if tr.Duration > 0 {
		since := time.Now().Add(-tr.Duration).UTC().Format(time.RFC3339)
		timeFilter = "@timestamp:[" + since + " TO *]"
	}
	if timeFilter != "" {
		if q == "" {
			q = timeFilter
		} else {
			q = timeFilter + " AND (" + q + ")"
		}
	}
	return q
}

func (m *Model) executeSearch() tea.Cmd {
	idx, cols := m.index, m.columns
	q := m.buildFullQuery()
	page := m.page
	var sa []interface{}
	if page > 0 && page-1 < len(m.pageSorts) {
		sa = m.pageSorts[page-1]
	}
	return func() tea.Msg {
		result, err := m.client.SearchDocs(idx, q, cols, pageSize, sa)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return SearchResultMsg{Result: result}
	}
}

func (m *Model) Update(msg tea.Msg) (views.View, tea.Cmd) {
	switch msg := msg.(type) {
	case IndicesLoadedMsg:
		m.indices = msg.Names
		m.filteredIdx = msg.Names
		m.loading = false
		m.err = nil
		return m, nil

	case MappingLoadedMsg:
		m.allFields = msg.Fields
		m.columns = defaultColumns(msg.Fields)
		m.columnSelected = make(map[string]bool)
		for _, c := range m.columns {
			m.columnSelected[c] = true
		}
		m.loading = false
		m.rebuildTable()
		return m, m.executeSearch()

	case SearchResultMsg:
		m.loading = false
		m.err = nil
		r := msg.Result
		m.total = r.Total
		m.hits = r.Hits
		// Store search_after for next page boundary
		if len(r.Hits) > 0 {
			lastHitSort := r.Hits[len(r.Hits)-1].Sort
			// Grow pageSorts to current page
			for len(m.pageSorts) <= m.page {
				m.pageSorts = append(m.pageSorts, nil)
			}
			m.pageSorts[m.page] = lastHitSort
		}
		m.updateTableRows()
		if m.followInterval > 0 {
			return m, followTickCmd(m.followInterval)
		}
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		m.loading = false
		return m, nil

	case FollowTickMsg:
		if m.followInterval > 0 && m.index != "" {
			return m, m.executeSearch()
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Route non-key messages to active text inputs
	if m.editing {
		var cmd tea.Cmd
		m.queryInput, cmd = m.queryInput.Update(msg)
		return m, cmd
	}
	if m.indexSearching {
		var cmd tea.Cmd
		m.indexInput, cmd = m.indexInput.Update(msg)
		return m, cmd
	}
	if m.pickingTime && m.timeCustomActive {
		var cmd tea.Cmd
		m.timeCustomInput, cmd = m.timeCustomInput.Update(msg)
		return m, cmd
	}
	if m.timeRangeActive {
		var cmd tea.Cmd
		if m.timeRangeField == 0 {
			m.timeFromInput, cmd = m.timeFromInput.Update(msg)
		} else {
			m.timeToInput, cmd = m.timeToInput.Update(msg)
		}
		return m, cmd
	}
	if m.buildingQuery {
		var cmd tea.Cmd
		if m.qbSt == qbSelectField {
			m.qbFieldInput, cmd = m.qbFieldInput.Update(msg)
		} else if m.qbSt == qbEnterValue {
			m.qbValueInput, cmd = m.qbValueInput.Update(msg)
		}
		return m, cmd
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

// --- View interface (rendering in view.go, key handlers in update.go) ---

func (m *Model) Name() string { return "Query" }

func (m *Model) HelpGroups() []views.HelpGroup {
	return []views.HelpGroup{
		{
			Title: "Query",
			Bindings: []key.Binding{
				m.keys.Search, m.keys.RawQuery, m.keys.Follow, m.keys.TimeRange, m.keys.Columns, m.keys.Refresh, m.keys.NextPage, m.keys.PrevPage, m.keys.Expand,
			},
		},
		{
			Title: "General",
			Bindings: []key.Binding{
				m.keys.Back, m.keys.Help, m.keys.Quit,
			},
		},
	}
}

func (m *Model) IsInputMode() bool {
	return m.editing || m.indexSearching || m.buildingQuery || m.pickingTime || m.timeRangeActive || m.viewingDoc
}

func (m *Model) PopPendingAction() *views.PendingAction {
	a := m.pendingAction
	m.pendingAction = nil
	return a
}

func (m *Model) StatusInfo() string {
	parts := []string{}
	if m.index != "" {
		parts = append(parts, m.index)
	}
	if m.followInterval > 0 {
		parts = append(parts, fmt.Sprintf("live: %s", m.followInterval))
	} else if m.index != "" {
		parts = append(parts, "paused")
	}
	if len(parts) == 0 {
		return "selecting index"
	}
	return strings.Join(parts, " | ")
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.table.SetHeight(height - 5)
	m.rebuildTable()
	m.updateTableRows()
}
