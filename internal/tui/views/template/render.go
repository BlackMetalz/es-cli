package template

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kienlt/es-cli/internal/tui/theme"
)

func (m *Model) updateTable() {
	rows := make([]table.Row, len(m.filtered))
	for i, t := range m.filtered {
		rows[i] = table.Row{
			t.Name,
			t.IndexPatterns,
			rightAlign(t.Shards, m.colWidths[2]),
			rightAlign(t.Replicas, m.colWidths[3]),
			t.ILMPolicy,
		}
	}
	m.table.SetRows(rows)
}

func (m *Model) postProcessTable(tableView string) string {
	lines := strings.Split(tableView, "\n")

	var selectedName string
	cursor := m.table.Cursor()
	if cursor >= 0 && cursor < len(m.filtered) {
		selectedName = m.filtered[cursor].Name
	}

	truncatedName := selectedName
	if len(truncatedName) > m.colWidths[0]-1 && m.colWidths[0] > 1 {
		truncatedName = truncatedName[:m.colWidths[0]-1]
	}

	selectedDone := false
	for i, line := range lines {
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

var fixedColWidths = [5]int{30, 30, 8, 10, 20}

const (
	minNameColWidth = 10
	maxNameColWidth = 60
)

func (m *Model) updateColumnWidths() {
	if m.width <= 0 {
		return
	}

	m.colWidths = fixedColWidths

	// Compute name column width from actual data
	nameWidth := len("name") // at least as wide as the header
	for _, t := range m.filtered {
		if len(t.Name) > nameWidth {
			nameWidth = len(t.Name)
		}
	}
	nameWidth += 2 // small padding
	if nameWidth < minNameColWidth {
		nameWidth = minNameColWidth
	}
	if nameWidth > maxNameColWidth {
		nameWidth = maxNameColWidth
	}
	m.colWidths[0] = nameWidth

	// Shrink name column further if terminal is narrow
	total := 0
	for _, w := range m.colWidths {
		total += w
	}
	if total > m.width {
		m.colWidths[0] = m.colWidths[0] - (total - m.width)
		if m.colWidths[0] < minNameColWidth {
			m.colWidths[0] = minNameColWidth
		}
	}

	cols := []table.Column{
		{Title: "name", Width: m.colWidths[0]},
		{Title: "index_patterns", Width: m.colWidths[1]},
		{Title: "shards", Width: m.colWidths[2]},
		{Title: "replicas", Width: m.colWidths[3]},
		{Title: "ilm_policy", Width: m.colWidths[4]},
	}

	m.table.SetWidth(m.width)
	m.table.SetColumns(cols)
	if len(m.filtered) > 0 {
		m.updateTable()
	}
}
