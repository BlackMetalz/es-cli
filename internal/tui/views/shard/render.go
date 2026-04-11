package shard

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kienlt/es-cli/internal/tui/theme"
)

func (m *Model) updateTable() {
	rows := make([]table.Row, len(m.filtered))
	for i, s := range m.filtered {
		rows[i] = table.Row{
			s.Index,
			rightAlign(s.ShardN, m.colWidths[1]),
			s.PriRep,
			s.State,
			rightAlign(s.Docs, m.colWidths[4]),
			rightAlign(s.Store, m.colWidths[5]),
			s.IP,
			s.Node,
		}
	}
	m.table.SetRows(rows)
}

// postProcessTable colorizes shard state and highlights the selected row.
func (m *Model) postProcessTable(tableView string) string {
	lines := strings.Split(tableView, "\n")

	// Get the selected index name for highlighting
	var selectedIndex string
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.filtered) {
		selectedIndex = m.filtered[cursor].Index
	}

	// Truncated name as it appears in the table
	truncatedName := selectedIndex
	if len(truncatedName) > m.colWidths[0]-1 && m.colWidths[0] > 1 {
		truncatedName = truncatedName[:m.colWidths[0]-1]
	}

	selectedDone := false
	for i, line := range lines {
		// Colorize state values
		if strings.Contains(line, "STARTED") {
			lines[i] = strings.ReplaceAll(lines[i], "STARTED", theme.HealthGreenStyle.Render("STARTED"))
		}
		if strings.Contains(line, "RELOCATING") {
			lines[i] = strings.ReplaceAll(lines[i], "RELOCATING", theme.HealthYellowStyle.Render("RELOCATING"))
		}
		if strings.Contains(line, "INITIALIZING") {
			lines[i] = strings.ReplaceAll(lines[i], "INITIALIZING", theme.HealthYellowStyle.Render("INITIALIZING"))
		}
		if strings.Contains(line, "UNASSIGNED") {
			lines[i] = strings.ReplaceAll(lines[i], "UNASSIGNED", theme.HealthRedStyle.Render("UNASSIGNED"))
		}

		// Highlight the selected row's index name
		if !selectedDone && selectedIndex != "" {
			nameToFind := truncatedName
			if strings.Contains(line, selectedIndex) {
				nameToFind = selectedIndex
			}
			if strings.Contains(line, nameToFind) {
				highlighted := theme.TableSelectedStyle.Render(nameToFind)
				lines[i] = strings.Replace(lines[i], nameToFind, highlighted, 1)
				selectedDone = true
			}
		}
	}
	return strings.Join(lines, "\n")
}

func rightAlign(s string, width int) string {
	n := len(s)
	if n >= width {
		return s
	}
	return strings.Repeat(" ", width-n) + s
}

var fixedColWidths = [8]int{30, 8, 8, 14, 12, 12, 16, 20}

const (
	minIndexColWidth = 10
	maxIndexColWidth = 60
)

func (m *Model) updateColumnWidths() {
	if m.width <= 0 {
		return
	}

	m.colWidths = fixedColWidths

	// Compute index column width from actual data
	idxWidth := len("index")
	for _, s := range m.filtered {
		if len(s.Index) > idxWidth {
			idxWidth = len(s.Index)
		}
	}
	idxWidth += 2
	if idxWidth < minIndexColWidth {
		idxWidth = minIndexColWidth
	}
	if idxWidth > maxIndexColWidth {
		idxWidth = maxIndexColWidth
	}
	m.colWidths[0] = idxWidth

	// Shrink index column if terminal is narrow
	total := 0
	for _, w := range m.colWidths {
		total += w
	}
	if total > m.width {
		m.colWidths[0] = m.colWidths[0] - (total - m.width)
		if m.colWidths[0] < minIndexColWidth {
			m.colWidths[0] = minIndexColWidth
		}
	}

	cols := []table.Column{
		{Title: "index", Width: m.colWidths[0]},
		{Title: "shard", Width: m.colWidths[1]},
		{Title: "prirep", Width: m.colWidths[2]},
		{Title: "state", Width: m.colWidths[3]},
		{Title: "docs", Width: m.colWidths[4]},
		{Title: "store", Width: m.colWidths[5]},
		{Title: "ip", Width: m.colWidths[6]},
		{Title: "node", Width: m.colWidths[7]},
	}

	m.table.SetWidth(m.width)
	m.table.SetColumns(cols)
	if len(m.filtered) > 0 {
		m.updateTable()
	}
}
