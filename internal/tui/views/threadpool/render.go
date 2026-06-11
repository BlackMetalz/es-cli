package threadpool

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/kienlt/es-cli/internal/tui/theme"
)

// Column indices
const (
	colNode     = 0
	colName     = 1
	colType     = 2
	colActive   = 3
	colSize     = 4
	colQueue    = 5
	colRejected = 6
	colLargest  = 7
)

var defaultColWidths = [8]int{20, 20, 8, 8, 6, 8, 10, 8}

const (
	minNodeColWidth = 10
	maxNodeColWidth = 40
	minNameColWidth = 10
	maxNameColWidth = 30
)

func (m *Model) updateTable() {
	rows := make([]table.Row, len(m.filtered))
	for i, tp := range m.filtered {
		sizeStr := fmt.Sprintf("%d", tp.Size)
		if tp.Size <= 0 {
			sizeStr = "-"
		}
		rows[i] = table.Row{
			tp.Node,
			tp.Name,
			tp.Type,
			rightAlign(fmt.Sprintf("%d", tp.Active), m.colWidths[colActive]),
			rightAlign(sizeStr, m.colWidths[colSize]),
			rightAlign(fmt.Sprintf("%d", tp.Queue), m.colWidths[colQueue]),
			rightAlign(fmt.Sprintf("%d", tp.Rejected), m.colWidths[colRejected]),
			rightAlign(fmt.Sprintf("%d", tp.Largest), m.colWidths[colLargest]),
		}
	}
	m.table.SetRows(rows)
}

// postProcessTable colorizes rejected/queue/active values and highlights the selected row.
// Identification of data lines is done by extracting node+name from fixed column positions
// and looking them up in a pre-built map; header/separator lines produce no match and are skipped.
func (m *Model) postProcessTable(tableView string) string {
	type poolStats struct {
		active, size, queue, rejected int
	}
	poolMap := make(map[string]poolStats, len(m.filtered))
	for _, tp := range m.filtered {
		poolMap[tp.Node+"\x00"+tp.Name] = poolStats{tp.Active, tp.Size, tp.Queue, tp.Rejected}
	}

	lines := strings.Split(tableView, "\n")

	cursor := m.table.Cursor()
	var selNode, selName string
	if cursor >= 0 && cursor < len(m.filtered) {
		sel := m.filtered[cursor]
		selNode = sel.Node
		selName = sel.Name
		if len(selName) > m.colWidths[colName]-1 && m.colWidths[colName] > 1 {
			selName = selName[:m.colWidths[colName]-1]
		}
	}
	selectedDone := false

	col0 := m.colWidths[colNode]
	col1 := m.colWidths[colName]

	for i, line := range lines {
		if len(line) < col0+col1 {
			continue
		}

		nodeStr := strings.TrimSpace(line[:col0])
		nameStr := strings.TrimSpace(line[col0 : col0+col1])
		stats, ok := poolMap[nodeStr+"\x00"+nameStr]

		if ok {
			if stats.rejected > 0 {
				lines[i] = replaceRightAligned(lines[i], stats.rejected, m.colWidths[colRejected],
					theme.HealthRedStyle.Render)
			}
			if stats.queue > 0 {
				lines[i] = replaceRightAligned(lines[i], stats.queue, m.colWidths[colQueue],
					theme.HealthYellowStyle.Render)
			}
			if stats.active > 0 && stats.size > 0 && stats.active >= stats.size {
				lines[i] = replaceRightAligned(lines[i], stats.active, m.colWidths[colActive],
					theme.HealthYellowStyle.Render)
			}
		}

		if !selectedDone && selName != "" && selNode != "" {
			if strings.Contains(line, selNode) && strings.Contains(line, selName) {
				lines[i] = strings.Replace(lines[i], selName, theme.TableSelectedStyle.Render(selName), 1)
				selectedDone = true
			}
		}
	}

	return strings.Join(lines, "\n")
}

// replaceRightAligned replaces the right-aligned rendering of n (padded to colWidth)
// with a colored version of just the numeric part.
func replaceRightAligned(line string, n, colWidth int, colorFn func(...string) string) string {
	numStr := fmt.Sprintf("%d", n)
	padding := strings.Repeat(" ", colWidth-len(numStr))
	plain := padding + numStr
	colored := padding + colorFn(numStr)
	return strings.Replace(line, plain, colored, 1)
}

func rightAlign(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

func (m *Model) updateColumnWidths() {
	if m.width <= 0 {
		return
	}

	m.colWidths = defaultColWidths

	nodeWidth := len("node")
	for _, tp := range m.filtered {
		if len(tp.Node) > nodeWidth {
			nodeWidth = len(tp.Node)
		}
	}
	nodeWidth += 2
	if nodeWidth < minNodeColWidth {
		nodeWidth = minNodeColWidth
	}
	if nodeWidth > maxNodeColWidth {
		nodeWidth = maxNodeColWidth
	}
	m.colWidths[colNode] = nodeWidth

	nameWidth := len("name")
	for _, tp := range m.filtered {
		if len(tp.Name) > nameWidth {
			nameWidth = len(tp.Name)
		}
	}
	nameWidth += 2
	if nameWidth < minNameColWidth {
		nameWidth = minNameColWidth
	}
	if nameWidth > maxNameColWidth {
		nameWidth = maxNameColWidth
	}
	m.colWidths[colName] = nameWidth

	total := 0
	for _, w := range m.colWidths {
		total += w
	}
	if total > m.width {
		m.colWidths[colName] -= total - m.width
		if m.colWidths[colName] < minNameColWidth {
			m.colWidths[colName] = minNameColWidth
		}
	}

	cols := []table.Column{
		{Title: "node", Width: m.colWidths[colNode]},
		{Title: "name", Width: m.colWidths[colName]},
		{Title: "type", Width: m.colWidths[colType]},
		{Title: "active", Width: m.colWidths[colActive]},
		{Title: "size", Width: m.colWidths[colSize]},
		{Title: "queue", Width: m.colWidths[colQueue]},
		{Title: "rejected", Width: m.colWidths[colRejected]},
		{Title: "largest", Width: m.colWidths[colLargest]},
	}

	m.table.SetWidth(m.width)
	m.table.SetColumns(cols)
	if len(m.filtered) > 0 {
		m.updateTable()
	}
}
