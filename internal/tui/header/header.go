package header

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kienlt/es-cli/internal/tui/theme"
	"github.com/kienlt/es-cli/internal/tui/views"
)

type Model struct {
	ClusterURL    string
	ClusterName   string
	ClusterHealth string
	ESVersion     string
	User          string
	ViewName      string
	HelpGroups    []views.HelpGroup
	Width         int
}

func New(clusterURL string) Model {
	return Model{
		ClusterURL: clusterURL,
		ViewName:   "Indices",
	}
}

var (
	labelStyle = lipgloss.NewStyle().
			Foreground(theme.ColorCyan).
			Bold(true)

	valueStyle = lipgloss.NewStyle().
			Foreground(theme.ColorWhite)

	groupTitleStyle = lipgloss.NewStyle().
			Foreground(theme.ColorYellow).
			Bold(true)
)

// Height returns the current header height in lines.
func (m Model) Height() int {
	helpRows := m.maxHelpRows()
	rows := helpRows + 1 // +1 for group title row
	if rows < 4 {
		rows = 4
	}
	return rows + 1 // +1 for separator
}

func (m Model) maxHelpRows() int {
	max := 0
	for _, g := range m.HelpGroups {
		if len(g.Bindings) > max {
			max = len(g.Bindings)
		}
	}
	return max
}

func (m Model) View() string {
	if m.Width <= 0 {
		return ""
	}

	clusterName := m.ClusterName
	if clusterName == "" {
		clusterName = "n/a"
	}
	esVersion := m.ESVersion
	if esVersion == "" {
		esVersion = "n/a"
	}
	user := m.User
	if user == "" {
		user = "n/a"
	}

	health := m.ClusterHealth
	if health == "" {
		health = "n/a"
	}
	healthStyled := theme.HealthStyle(health).Render(health)

	infoLines := []string{
		labelStyle.Render("URL:     ") + valueStyle.Render(m.ClusterURL),
		labelStyle.Render("Cluster: ") + valueStyle.Render(clusterName),
		labelStyle.Render("Health:  ") + healthStyled,
		labelStyle.Render("User:    ") + valueStyle.Render(user),
		labelStyle.Render("ES Rev:  ") + valueStyle.Render(esVersion),
	}

	// Info on left, grouped help columns on right
	helpColumns := m.buildGroupColumns()

	// Determine total rows needed (group title + max bindings)
	helpRows := m.maxHelpRows() + 1 // +1 for title
	totalRows := helpRows
	if totalRows < len(infoLines) {
		totalRows = len(infoLines)
	}

	// Pad info lines
	for len(infoLines) < totalRows {
		infoLines = append(infoLines, "")
	}
	// Pad help columns
	for i := range helpColumns {
		for len(helpColumns[i]) < totalRows {
			helpColumns[i] = append(helpColumns[i], "")
		}
	}

	leftWidth := 40
	var merged []string
	for row := 0; row < totalRows; row++ {
		left := infoLines[row]
		leftVisible := lipgloss.Width(left)
		padding := leftWidth - leftVisible
		if padding < 0 {
			padding = 0
		}

		var right string
		for _, col := range helpColumns {
			right += col[row]
		}

		merged = append(merged, left+strings.Repeat(" ", padding)+right)
	}

	sep := theme.SeparatorStyle.Render(strings.Repeat("─", m.Width))
	merged = append(merged, sep)

	return strings.Join(merged, "\n")
}

func (m Model) buildGroupColumns() [][]string {
	if len(m.HelpGroups) == 0 {
		return nil
	}

	maxRows := m.maxHelpRows() + 1 // +1 for title

	// Calculate column width for each group
	var columns [][]string
	for _, g := range m.HelpGroups {
		// Find widest entry in this group
		colWidth := len(g.Title) + 2
		for _, b := range g.Bindings {
			h := b.Help()
			w := len(fmt.Sprintf("<%s>", h.Key)) + 1 + len(h.Desc) + 3
			if w > colWidth {
				colWidth = w
			}
		}
		if colWidth > 35 {
			colWidth = 35
		}

		lines := make([]string, maxRows)

		// Title row
		title := groupTitleStyle.Render(g.Title)
		titleVisible := lipgloss.Width(title)
		titlePad := colWidth - titleVisible
		if titlePad < 0 {
			titlePad = 0
		}
		lines[0] = title + strings.Repeat(" ", titlePad)

		// Binding rows
		for i, b := range g.Bindings {
			h := b.Help()
			keyStr := theme.HelpKeyStyle.Render(fmt.Sprintf("<%s>", h.Key))
			descStr := theme.HelpDescStyle.Render(" " + h.Desc)
			cell := keyStr + descStr

			cellVisible := lipgloss.Width(cell)
			pad := colWidth - cellVisible
			if pad < 0 {
				pad = 0
			}
			lines[i+1] = cell + strings.Repeat(" ", pad)
		}

		// Pad remaining rows
		for i := len(g.Bindings) + 1; i < maxRows; i++ {
			lines[i] = strings.Repeat(" ", colWidth)
		}

		columns = append(columns, lines)
	}

	return columns
}
