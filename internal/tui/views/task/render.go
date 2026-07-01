package task

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/kienlt/es-cli/internal/es"
	"github.com/kienlt/es-cli/internal/tui/theme"
)

const (
	colDuration    = 0
	colNode        = 1
	colAction      = 2
	colType        = 3
	colCancellable = 4
	colDescription = 5
)

const topN = 20

var defaultColWidths = [6]int{10, 20, 35, 12, 10, 20}

const (
	minNodeWidth   = 8
	maxNodeWidth   = 30
	minActionWidth = 20
	maxActionWidth = 50
	minDescWidth   = 10
)

func (m *Model) updateTable() {
	rows := make([]table.Row, len(m.filtered))
	for i, t := range m.filtered {
		cancelStr := "no"
		if t.Cancellable {
			cancelStr = "yes"
		}
		rows[i] = table.Row{
			rightAlign(es.FormatTaskDuration(t.RunningTimeNanos), m.colWidths[colDuration]),
			t.NodeName,
			t.Action,
			t.Type,
			cancelStr,
			t.Description,
		}
	}
	m.table.SetRows(rows)
}

// cursorSentinel is the ANSI sequence bubbles emits around the cursor row when
// Styles.Selected = NewStyle().Reverse(true). We strip it and replace with our
// own highlight; matching by sentinel is robust against duplicate row content
// (same node+action across many tasks) and viewport scrolling.
const cursorSentinel = "\x1b[7m"
const ansiReset = "\x1b[0m"

func (m *Model) postProcessTable(tableView string) string {
	lines := strings.Split(tableView, "\n")

	cursor := m.table.Cursor()
	var selAction string
	if cursor >= 0 && cursor < len(m.filtered) {
		selAction = m.filtered[cursor].Action
		if len(selAction) > m.colWidths[colAction]-1 && m.colWidths[colAction] > 1 {
			selAction = selAction[:m.colWidths[colAction]-1]
		}
	}

	col0 := m.colWidths[colDuration]

	for i, line := range lines {
		isCursor := strings.HasPrefix(line, cursorSentinel)
		if isCursor {
			line = strings.TrimSuffix(strings.TrimPrefix(line, cursorSentinel), ansiReset)
			lines[i] = line
		}

		if len(line) < col0 {
			continue
		}
		durStr := strings.TrimSpace(line[:col0])
		if durStr == "" {
			continue
		}

		if col, ok := durationColor(durStr); ok {
			padding := strings.Repeat(" ", col0-len(durStr))
			lines[i] = strings.Replace(lines[i], padding+durStr, padding+col.Render(durStr), 1)
		}

		if isCursor && selAction != "" {
			lines[i] = strings.Replace(lines[i], selAction, theme.TableSelectedStyle.Render(selAction), 1)
		}
	}

	return strings.Join(lines, "\n")
}

// durationColor returns the color for a formatted duration string.
// Red for >= 1 minute, yellow for >= 10 seconds.
func durationColor(durStr string) (lipgloss.Style, bool) {
	// >= 1 minute: "Xm YYs" format — has "m" but doesn't end with "ms"
	if strings.Contains(durStr, "m") && !strings.HasSuffix(durStr, "ms") {
		return theme.HealthRedStyle, true
	}
	// >= 10 seconds: plain seconds "XX.Xs" format
	if strings.HasSuffix(durStr, "s") &&
		!strings.HasSuffix(durStr, "ms") &&
		!strings.HasSuffix(durStr, "µs") &&
		!strings.HasSuffix(durStr, "ns") {
		numStr := strings.TrimSuffix(durStr, "s")
		if f, err := strconv.ParseFloat(numStr, 64); err == nil && f >= 10 {
			return theme.HealthYellowStyle, true
		}
	}
	return lipgloss.Style{}, false
}

func (m *Model) updateColumnWidths() {
	if m.width <= 0 {
		return
	}

	m.colWidths = defaultColWidths

	nodeWidth := len("node")
	for _, t := range m.filtered {
		if len(t.NodeName) > nodeWidth {
			nodeWidth = len(t.NodeName)
		}
	}
	nodeWidth += 2
	if nodeWidth < minNodeWidth {
		nodeWidth = minNodeWidth
	}
	if nodeWidth > maxNodeWidth {
		nodeWidth = maxNodeWidth
	}
	m.colWidths[colNode] = nodeWidth

	actionWidth := len("action")
	for _, t := range m.filtered {
		if len(t.Action) > actionWidth {
			actionWidth = len(t.Action)
		}
	}
	actionWidth += 2
	if actionWidth < minActionWidth {
		actionWidth = minActionWidth
	}
	if actionWidth > maxActionWidth {
		actionWidth = maxActionWidth
	}
	m.colWidths[colAction] = actionWidth

	fixed := 0
	for i := 0; i < colDescription; i++ {
		fixed += m.colWidths[i]
	}
	descWidth := m.width - fixed
	if descWidth < minDescWidth {
		descWidth = minDescWidth
	}
	m.colWidths[colDescription] = descWidth

	cols := []table.Column{
		{Title: "duration", Width: m.colWidths[colDuration]},
		{Title: "node", Width: m.colWidths[colNode]},
		{Title: "action", Width: m.colWidths[colAction]},
		{Title: "type", Width: m.colWidths[colType]},
		{Title: "cancelable", Width: m.colWidths[colCancellable]},
		{Title: "description", Width: m.colWidths[colDescription]},
	}

	m.table.SetWidth(m.width)
	m.table.SetColumns(cols)
	if len(m.filtered) > 0 {
		m.updateTable()
	}
}

func rightAlign(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}
