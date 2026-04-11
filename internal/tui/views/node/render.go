package node

import (
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kienlt/es-cli/internal/tui/theme"
)

var defaultColWidths = [11]int{16, 14, 7, 7, 5, 9, 9, 9, 14, 8, 7}

const (
	minNameColWidth = 10
	maxNameColWidth = 40
)

func (m *Model) updateTable() {
	rows := make([]table.Row, len(m.filtered))
	for i, n := range m.filtered {
		rows[i] = table.Row{
			n.Name,
			n.IP,
			rightAlign(n.HeapPercent, m.colWidths[2]),
			rightAlign(n.RAMPercent, m.colWidths[3]),
			rightAlign(n.CPU, m.colWidths[4]),
			rightAlign(n.Load1m, m.colWidths[5]),
			rightAlign(n.Load5m, m.colWidths[6]),
			rightAlign(n.Load15m, m.colWidths[7]),
			n.NodeRole,
			n.Master,
			rightAlign(n.DiskUsedPercent, m.colWidths[10]),
		}
	}
	m.table.SetRows(rows)
}

// postProcessTable colorizes high-usage values and highlights the selected row.
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
		// Colorize numeric values > 80 as red, > 60 as yellow
		lines[i] = colorizeValues(line)

		// Highlight master node indicator
		if strings.Contains(line, "*") {
			lines[i] = strings.Replace(lines[i], "*", theme.HealthGreenStyle.Render("*"), 1)
		}

		// Highlight selected node name
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

// colorizeValues finds numeric tokens in table cells and applies red/yellow coloring.
func colorizeValues(line string) string {
	// Split by multiple spaces to find cell boundaries
	tokens := strings.Fields(line)
	for _, tok := range tokens {
		val := parsePercent(tok)
		if val <= 0 {
			continue
		}
		var colored string
		if val > 80 {
			colored = theme.HealthRedStyle.Render(tok)
		} else if val > 60 {
			colored = theme.HealthYellowStyle.Render(tok)
		} else {
			continue
		}
		line = strings.Replace(line, tok, colored, 1)
	}
	return line
}

func rightAlign(s string, width int) string {
	n := len(s)
	if n >= width {
		return s
	}
	return strings.Repeat(" ", width-n) + s
}

func (m *Model) updateColumnWidths() {
	if m.width <= 0 {
		return
	}

	m.colWidths = defaultColWidths

	// Compute name column width from actual data
	nameWidth := len("name")
	for _, n := range m.filtered {
		if len(n.Name) > nameWidth {
			nameWidth = len(n.Name)
		}
	}
	nameWidth += 2
	if nameWidth < minNameColWidth {
		nameWidth = minNameColWidth
	}
	if nameWidth > maxNameColWidth {
		nameWidth = maxNameColWidth
	}
	m.colWidths[0] = nameWidth

	// Shrink name column if terminal is narrow
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
		{Title: "ip", Width: m.colWidths[1]},
		{Title: "heap%", Width: m.colWidths[2]},
		{Title: "ram%", Width: m.colWidths[3]},
		{Title: "cpu", Width: m.colWidths[4]},
		{Title: "load_1m", Width: m.colWidths[5]},
		{Title: "load_5m", Width: m.colWidths[6]},
		{Title: "load_15m", Width: m.colWidths[7]},
		{Title: "role", Width: m.colWidths[8]},
		{Title: "master", Width: m.colWidths[9]},
		{Title: "disk%", Width: m.colWidths[10]},
	}

	m.table.SetWidth(m.width)
	m.table.SetColumns(cols)
	if len(m.filtered) > 0 {
		m.updateTable()
	}
}
