package dashboard

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kienlt/es-cli/internal/es"
	"github.com/kienlt/es-cli/internal/tui/theme"
	"github.com/kienlt/es-cli/internal/tui/views"
)

type DashboardLoadedMsg struct {
	Data *es.DashboardData
}

type ErrorMsg struct {
	Err error
}

type Model struct {
	client     *es.Client
	keys       KeyMap
	data       *es.DashboardData
	width      int
	height     int
	loading    bool
	err        error
	showHidden bool // include dot-indices in Index Analyze section
}

var _ views.View = (*Model)(nil)

func New(client *es.Client) *Model {
	return &Model{
		client:  client,
		keys:    DefaultKeyMap(),
		loading: true,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.fetchDashboard()
}

func (m *Model) fetchDashboard() tea.Cmd {
	return func() tea.Msg {
		data, err := m.client.GetDashboardData()
		if err != nil {
			return ErrorMsg{Err: err}
		}
		return DashboardLoadedMsg{Data: data}
	}
}

func (m *Model) Update(msg tea.Msg) (views.View, tea.Cmd) {
	switch msg := msg.(type) {
	case DashboardLoadedMsg:
		m.data = msg.Data
		m.loading = false
		m.err = nil
		return m, nil

	case ErrorMsg:
		m.err = msg.Err
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (views.View, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Refresh):
		m.loading = true
		return m, m.fetchDashboard()
	case key.Matches(msg, m.keys.ToggleHidden):
		m.showHidden = !m.showHidden
		return m, nil
	case key.Matches(msg, m.keys.Help):
		return m, nil
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	}
	return m, nil
}

// --- Rendering ---

var (
	sectionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.ColorBlue).
			Padding(1, 2)

	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.ColorCyan)

	labelStyle = lipgloss.NewStyle().
			Foreground(theme.ColorWhite).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(theme.ColorWhite)
)

func (m *Model) View() string {
	if m.loading {
		return "\n  Loading dashboard..."
	}
	if m.err != nil {
		return "\n  " + theme.ErrorStyle.Render("Error: "+m.err.Error())
	}

	d := m.data

	// Overview section
	healthDot := theme.HealthStyle(d.Health).Render("●")
	healthDesc := d.HealthDescription
	if healthDesc == "" {
		healthDesc = d.Health
	}
	license := d.License
	if license == "" {
		license = "n/a"
	}
	overview := renderSection("Overview", []row{
		{"Health", healthDot + " " + healthDesc, ""},
		{"Version", d.Version, ""},
		{"Uptime", formatUptime(d.Uptime), ""},
		{"License", theme.HelpKeyStyle.Render(license), ""},
	})

	// Nodes section — show percentage as main value, raw bytes as subtitle
	diskPct := fmt.Sprintf("%.2f%%", d.DiskAvailPercent())
	diskDetail := es.FormatBytes(fmt.Sprintf("%.0f", d.DiskAvailBytes)) + " / " + es.FormatBytes(fmt.Sprintf("%.0f", d.DiskTotalBytes))
	heapPct := fmt.Sprintf("%.2f%%", d.HeapUsedPercent())
	heapDetail := es.FormatBytes(fmt.Sprintf("%.0f", d.HeapUsedBytes)) + " / " + es.FormatBytes(fmt.Sprintf("%.0f", d.HeapMaxBytes))

	nodes := renderSection(fmt.Sprintf("Nodes: %d", d.NodeCount), []row{
		{"Disk Available", diskPct, diskDetail},
		{"JVM Heap", heapPct, heapDetail},
	})

	// Indices section
	indices := renderSection(fmt.Sprintf("Indices: %d", d.IndexCount), []row{
		{"Documents", formatNumber(d.DocCount), ""},
		{"Disk Usage", d.DiskUsage, ""},
		{"Primary Shards", fmt.Sprintf("%d", d.PrimaryShards), ""},
		{"Replica Shards", fmt.Sprintf("%d", d.ReplicaShards), ""},
	})

	// Cluster Health section — active % first for easy monitoring
	pctStr := fmt.Sprintf("%.2f%%", d.ActiveShardsPercent)
	clusterHealth := renderSection("Cluster Health", []row{
		{"Active %", activePctStyle(d.ActiveShardsPercent).Render(pctStr), fmt.Sprintf("%d / %d shards", d.ActiveShards, d.ActiveShards+d.UnassignedShards+d.InitializingShards)},
		{"Relocating", shardCountValue(d.RelocatingShards, theme.HealthYellowStyle), ""},
		{"Initializing", shardCountValue(d.InitializingShards, theme.HealthYellowStyle), ""},
		{"Unassigned", shardCountValue(d.UnassignedShards, theme.HealthRedStyle), ""},
		{"Delayed", shardCountValue(d.DelayedUnassigned, theme.HealthYellowStyle), ""},
		{"In-Flight", shardCountValue(d.InFlightFetch, theme.HealthYellowStyle), ""},
		{"Pending Tasks", shardCountValue(d.PendingTasks, theme.HealthYellowStyle), ""},
		{"Task Wait", formatTaskWait(d.TaskMaxWaitMs), ""},
	})

	// Layout: 4 boxes side-by-side when very wide, else 3 boxes + cluster health row below
	var topContent string
	if m.width >= 120 {
		boxWidth := (m.width - 6) / 4

		s := sectionStyle.Width(boxWidth)
		box1 := s.Render(overview)
		box2 := s.Render(nodes)
		box3 := s.Render(indices)
		box4 := s.Render(clusterHealth)

		h1 := lipgloss.Height(box1)
		h2 := lipgloss.Height(box2)
		h3 := lipgloss.Height(box3)
		h4 := lipgloss.Height(box4)
		maxH := h1
		for _, h := range []int{h2, h3, h4} {
			if h > maxH {
				maxH = h
			}
		}

		styledOverview := s.Render(overview + strings.Repeat("\n", maxH-h1))
		styledNodes := s.Render(nodes + strings.Repeat("\n", maxH-h2))
		styledIndices := s.Render(indices + strings.Repeat("\n", maxH-h3))
		styledHealth := s.Render(clusterHealth + strings.Repeat("\n", maxH-h4))
		topContent = lipgloss.JoinHorizontal(lipgloss.Top, styledOverview, styledNodes, styledIndices, styledHealth)
	} else if m.width >= 80 {
		boxWidth := (m.width - 6) / 3

		s := sectionStyle.Width(boxWidth)
		box1 := s.Render(overview)
		box2 := s.Render(nodes)
		box3 := s.Render(indices)

		h1 := lipgloss.Height(box1)
		h2 := lipgloss.Height(box2)
		h3 := lipgloss.Height(box3)
		maxH := h1
		if h2 > maxH {
			maxH = h2
		}
		if h3 > maxH {
			maxH = h3
		}

		styledOverview := s.Render(overview + strings.Repeat("\n", maxH-h1))
		styledNodes := s.Render(nodes + strings.Repeat("\n", maxH-h2))
		styledIndices := s.Render(indices + strings.Repeat("\n", maxH-h3))
		row1 := lipgloss.JoinHorizontal(lipgloss.Top, styledOverview, styledNodes, styledIndices)
		row2 := sectionStyle.Width(m.width - 4).Render(clusterHealth)
		topContent = row1 + "\n" + row2
	} else {
		w := m.width - 4
		topContent = lipgloss.JoinVertical(lipgloss.Left,
			sectionStyle.Width(w).Render(overview),
			sectionStyle.Width(w).Render(nodes),
			sectionStyle.Width(w).Render(indices),
			sectionStyle.Width(w).Render(clusterHealth),
		)
	}

	// Index Analyze section: one wide box, content wraps into columns before hitting bottom
	topH := lipgloss.Height(topContent)
	analyzeBox := m.renderAnalyzeBox(topH)

	result := "\n" + topContent + "\n" + analyzeBox

	// Pad to fill available height so status bar sticks to bottom
	contentHeight := lipgloss.Height(result)
	if contentHeight < m.height {
		result += strings.Repeat("\n", m.height-contentHeight)
	}

	return result
}

// renderAnalyzeBox renders the Index Analyze section as a single wide box.
// When the pattern list would overflow the available height, it flows into
// additional columns inside the same box (newspaper-column style).
func (m *Model) renderAnalyzeBox(topH int) string {
	patterns := m.visiblePatterns()

	// How many pattern rows fit in one column before hitting the bottom.
	// Layout chain: 1 (\n before content) + topH + 1 (\n between rows) + analyzeBox = m.height
	// analyzeBox chrome: 1 top-border + 1 top-pad + 1 bot-pad + 1 bot-border = 4 lines
	// Fixed content inside box: title(1) + blank(1) + table-header(1) + separator(1) = 4 lines
	patternRowsPerCol := m.height - 1 - topH - 1 - 4 - 4
	if patternRowsPerCol < 1 {
		patternRowsPerCol = 1
	}

	numCols := 1
	if len(patterns) > patternRowsPerCol {
		numCols = 2
	}
	if len(patterns) > 2*patternRowsPerCol {
		numCols = 3
	}

	// Inner content width of the full-width box.
	// sectionStyle border = 1 each side (2), padding = 2 each side (4) → 6 total overhead.
	boxW := m.width - 4
	contentW := boxW - 6
	colW := contentW / numCols

	// Title + toggle hint
	hiddenLabel := subtitleStyle.Render("h: show hidden")
	if m.showHidden {
		hiddenLabel = theme.HealthYellowStyle.Render("h: hide hidden")
	}
	title := sectionTitleStyle.Render("Index Analyze") + "  " + hiddenLabel

	var body string
	if numCols == 1 {
		body = renderPatternTable(patterns, contentW)
	} else {
		// Distribute patterns into columns, filling top-to-bottom then next column.
		styledCols := make([]string, numCols)
		maxTableH := 0
		for c := 0; c < numCols; c++ {
			start := c * patternRowsPerCol
			if start >= len(patterns) {
				styledCols[c] = lipgloss.NewStyle().Width(colW).Render("")
				continue
			}
			end := start + patternRowsPerCol
			if end > len(patterns) {
				end = len(patterns)
			}
			tbl := renderPatternTable(patterns[start:end], colW)
			if h := lipgloss.Height(tbl); h > maxTableH {
				maxTableH = h
			}
			styledCols[c] = tbl
		}
		// Pad shorter columns so JoinHorizontal aligns cleanly
		for i, col := range styledCols {
			pad := strings.Repeat("\n", maxTableH-lipgloss.Height(col))
			styledCols[i] = lipgloss.NewStyle().Width(colW).Render(col + pad)
		}
		body = lipgloss.JoinHorizontal(lipgloss.Top, styledCols...)
	}

	content := title + "\n\n" + body
	return sectionStyle.Width(boxW).Render(content)
}

// visiblePatterns returns PatternStats filtered by the showHidden toggle.
func (m *Model) visiblePatterns() []es.IndexPatternStat {
	all := m.data.PatternStats
	if m.showHidden {
		return all
	}
	out := make([]es.IndexPatternStat, 0, len(all))
	for _, ps := range all {
		if !strings.HasPrefix(ps.Pattern, ".") {
			out = append(out, ps)
		}
	}
	return out
}

// renderPatternTable renders a list of patterns as a table with innerWidth content width.
func renderPatternTable(patterns []es.IndexPatternStat, innerWidth int) string {
	if len(patterns) == 0 {
		return ""
	}
	const (
		colIdx    = 5
		colShards = 8
		colDisk   = 10
	)
	colPattern := innerWidth - colIdx - colShards - colDisk
	if colPattern < 10 {
		colPattern = 10
	}

	var b strings.Builder
	b.WriteString(
		labelStyle.Width(colPattern).Render("Pattern") +
			labelStyle.Width(colIdx).Render("Idx") +
			labelStyle.Width(colShards).Render("Shards") +
			labelStyle.Width(colDisk).Render("Disk") + "\n",
	)
	b.WriteString(subtitleStyle.Render(strings.Repeat("─", colPattern+colIdx+colShards+colDisk)) + "\n")

	for _, ps := range patterns {
		pattern := ps.Pattern
		if len(pattern) > colPattern-1 {
			pattern = pattern[:colPattern-4] + "..."
		}
		disk := es.FormatBytes(fmt.Sprintf("%d", ps.DiskBytes))
		b.WriteString(
			valueStyle.Width(colPattern).Render(pattern) +
				subtitleStyle.Width(colIdx).Render(fmt.Sprintf("%d", ps.IndexCount)) +
				subtitleStyle.Width(colShards).Render(fmt.Sprintf("%d", ps.Shards)) +
				valueStyle.Width(colDisk).Render(disk) + "\n",
		)
	}
	return b.String()
}

type row struct {
	label    string
	value    string
	subtitle string // dimmed text below value (e.g., raw bytes)
}

var subtitleStyle = lipgloss.NewStyle().Foreground(theme.ColorGray)

func shardCountValue(n int, warnStyle lipgloss.Style) string {
	if n == 0 {
		return valueStyle.Render("0")
	}
	return warnStyle.Render(fmt.Sprintf("%d", n))
}

func activePctStyle(pct float64) lipgloss.Style {
	if pct >= 100 {
		return theme.HealthGreenStyle
	}
	if pct >= 75 {
		return theme.HealthYellowStyle
	}
	return theme.HealthRedStyle
}

func formatTaskWait(ms int) string {
	if ms == 0 {
		return valueStyle.Render("0ms")
	}
	if ms >= 1000 {
		return theme.HealthYellowStyle.Render(fmt.Sprintf("%dms", ms))
	}
	return valueStyle.Render(fmt.Sprintf("%dms", ms))
}

func renderSection(title string, rows []row) string {
	var b strings.Builder
	b.WriteString(sectionTitleStyle.Render(title))
	b.WriteString("\n\n")

	// Fixed label column width for consistent alignment
	const labelWidth = 16

	for _, r := range rows {
		lbl := labelStyle.Width(labelWidth).Render(r.label)
		b.WriteString(lbl + valueStyle.Render(r.value) + "\n")
		if r.subtitle != "" {
			pad := strings.Repeat(" ", labelWidth)
			b.WriteString(pad + subtitleStyle.Render(r.subtitle) + "\n")
		}
	}
	return b.String()
}

func formatUptime(d time.Duration) string {
	if d <= 0 {
		return "n/a"
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	s := fmt.Sprintf("%d", n)
	var result []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			result = append(result, ',')
		}
		result = append(result, byte(c))
	}
	return string(result)
}

// --- Interface methods ---

func (m *Model) Name() string { return "Dashboard" }

func (m *Model) HelpGroups() []views.HelpGroup {
	return []views.HelpGroup{
		{
			Title: "Dashboard",
			Bindings: []key.Binding{
				m.keys.Refresh,
				m.keys.ToggleHidden,
				m.keys.Help,
				m.keys.Quit,
			},
		},
	}
}

func (m *Model) IsInputMode() bool                      { return false }
func (m *Model) PopPendingAction() *views.PendingAction { return nil }

func (m *Model) StatusInfo() string {
	if m.data != nil {
		return "cluster: " + m.data.Health
	}
	return ""
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}
