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
	client  *es.Client
	keys    KeyMap
	data    *es.DashboardData
	width   int
	height  int
	loading bool
	err     error
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

	// Layout: horizontal if wide enough, else vertical
	var content string
	if m.width >= 80 {
		boxWidth := (m.width - 6) / 3

		// Render all 3 boxes first, then find max height and re-render with equal height
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

		// Pad content to equalize rendered box heights
		pad1 := strings.Repeat("\n", maxH-h1)
		pad2 := strings.Repeat("\n", maxH-h2)
		pad3 := strings.Repeat("\n", maxH-h3)

		styledOverview := s.Render(overview + pad1)
		styledNodes := s.Render(nodes + pad2)
		styledIndices := s.Render(indices + pad3)
		content = lipgloss.JoinHorizontal(lipgloss.Top, styledOverview, styledNodes, styledIndices)
	} else {
		styledOverview := sectionStyle.Width(m.width - 4).Render(overview)
		styledNodes := sectionStyle.Width(m.width - 4).Render(nodes)
		styledIndices := sectionStyle.Width(m.width - 4).Render(indices)
		content = lipgloss.JoinVertical(lipgloss.Left, styledOverview, styledNodes, styledIndices)
	}

	result := "\n" + content

	// Pad to fill available height so status bar sticks to bottom
	contentHeight := lipgloss.Height(result)
	if contentHeight < m.height {
		result += strings.Repeat("\n", m.height-contentHeight)
	}

	return result
}

type row struct {
	label    string
	value    string
	subtitle string // dimmed text below value (e.g., raw bytes)
}

var subtitleStyle = lipgloss.NewStyle().Foreground(theme.ColorGray)

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
