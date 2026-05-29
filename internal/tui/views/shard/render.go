package shard

import (
	"regexp"
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
// The selected line is identified by matching multiple cells from the row data
// because a single index has many rows in this view (one per shard × prirep),
// so matching by index name alone would always land on the first row of the group.
func (m *Model) postProcessTable(tableView string) string {
	lines := strings.Split(tableView, "\n")

	selectedLine := m.findSelectedLine(lines)

	var selectedNameToHighlight string
	if cursor := m.table.Cursor(); cursor >= 0 && cursor < len(m.filtered) {
		selectedNameToHighlight = m.filtered[cursor].Index
		if len(selectedNameToHighlight) > m.colWidths[0]-1 && m.colWidths[0] > 1 {
			selectedNameToHighlight = selectedNameToHighlight[:m.colWidths[0]-1]
		}
	}

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

		if i == selectedLine && selectedNameToHighlight != "" {
			highlighted := theme.TableSelectedStyle.Render(selectedNameToHighlight)
			lines[i] = strings.Replace(lines[i], selectedNameToHighlight, highlighted, 1)
		}
	}
	return strings.Join(lines, "\n")
}

// findSelectedLine returns the index in lines that corresponds to the cursor's row,
// or -1 if not found. Matches on (index, shard, prirep) which is unique per row;
// node is appended when present to disambiguate replicas of the same shard.
func (m *Model) findSelectedLine(lines []string) int {
	selRow := m.table.SelectedRow()
	if len(selRow) < 8 {
		return -1
	}

	selIndex := selRow[0]
	if len(selIndex) > m.colWidths[0]-1 && m.colWidths[0] > 1 {
		selIndex = selIndex[:m.colWidths[0]-1]
	}
	selShard := strings.TrimSpace(selRow[1])
	selPrirep := strings.TrimSpace(selRow[2])
	selNode := strings.TrimSpace(selRow[7])

	pattern := regexp.QuoteMeta(selIndex) + `\s+` +
		regexp.QuoteMeta(selShard) + `\s+` +
		regexp.QuoteMeta(selPrirep) + `(\s|$)`
	if selNode != "" {
		pattern += `.*` + regexp.QuoteMeta(selNode)
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return -1
	}
	for i, line := range lines {
		if re.MatchString(line) {
			return i
		}
	}
	return -1
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
	minNodeColWidth  = 10
	maxNodeColWidth  = 60
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

	// Compute node column width from actual data (RELOCATING nodes are long)
	nodeWidth := len("node")
	for _, s := range m.filtered {
		if len(s.Node) > nodeWidth {
			nodeWidth = len(s.Node)
		}
	}
	nodeWidth += 2
	if nodeWidth < minNodeColWidth {
		nodeWidth = minNodeColWidth
	}
	if nodeWidth > maxNodeColWidth {
		nodeWidth = maxNodeColWidth
	}
	m.colWidths[7] = nodeWidth

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
