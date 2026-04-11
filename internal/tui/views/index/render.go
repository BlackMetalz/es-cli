package index

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kienlt/es-cli/internal/tui/theme"
)

func (m *Model) updateTable() {
	rows := make([]table.Row, len(m.filtered))
	for i, idx := range m.filtered {
		rows[i] = table.Row{
			idx.Health,
			idx.Status,
			idx.Name,
			rightAlign(idx.Pri, m.colWidths[3]),
			rightAlign(idx.Rep, m.colWidths[4]),
			rightAlign(idx.DocsCount, m.colWidths[5]),
			rightAlign(idx.StoreSize, m.colWidths[6]),
		}
	}
	m.table.SetRows(rows)
}

// postProcessTable colorizes health values and highlights the index name on the selected row.
// We handle selected row ourselves because the bubbles table's Selected style only covers
// the joined cells, not the full terminal width, and interferes with per-cell ANSI styling.
func (m *Model) postProcessTable(tableView string) string {
	lines := strings.Split(tableView, "\n")

	// Get the selected index name for highlighting
	var selectedName string
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.filtered) {
		selectedName = m.filtered[cursor].Name
	}

	// Truncated name as it appears in the table (for long names)
	truncatedName := selectedName
	if len(truncatedName) > m.colWidths[2]-1 && m.colWidths[2] > 1 {
		truncatedName = truncatedName[:m.colWidths[2]-1]
	}

	selectedDone := false
	for i, line := range lines {
		// Colorize health values
		for _, health := range []string{"green", "yellow", "red"} {
			if strings.Contains(line, health) {
				lines[i] = strings.ReplaceAll(lines[i], health, theme.HealthStyle(health).Render(health))
			}
		}

		// Highlight the selected index name (match by content, not line number)
		if !selectedDone && selectedName != "" {
			nameToFind := truncatedName
			if strings.Contains(line, selectedName) {
				nameToFind = selectedName
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

var fixedColWidths = [7]int{8, 8, 60, 15, 15, 12, 16}

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
	idxWidth := len("index") // at least as wide as the header
	for _, idx := range m.filtered {
		if len(idx.Name) > idxWidth {
			idxWidth = len(idx.Name)
		}
	}
	idxWidth += 2 // small padding
	if idxWidth < minIndexColWidth {
		idxWidth = minIndexColWidth
	}
	if idxWidth > maxIndexColWidth {
		idxWidth = maxIndexColWidth
	}
	m.colWidths[2] = idxWidth

	// Shrink index column further if terminal is narrow
	total := 0
	for _, w := range m.colWidths {
		total += w
	}
	if total > m.width {
		m.colWidths[2] = m.colWidths[2] - (total - m.width)
		if m.colWidths[2] < minIndexColWidth {
			m.colWidths[2] = minIndexColWidth
		}
	}

	cols := []table.Column{
		{Title: "health", Width: m.colWidths[0]},
		{Title: "status", Width: m.colWidths[1]},
		{Title: "index", Width: m.colWidths[2]},
		{Title: "primary shard", Width: m.colWidths[3]},
		{Title: "replica shard", Width: m.colWidths[4]},
		{Title: "docs.count", Width: m.colWidths[5]},
		{Title: "pri.store.size", Width: m.colWidths[6]},
	}

	m.table.SetWidth(m.width)
	m.table.SetColumns(cols)
	if len(m.filtered) > 0 {
		m.updateTable()
	}
}
