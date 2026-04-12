package query

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kienlt/es-cli/internal/es"
	"github.com/kienlt/es-cli/internal/tui/theme"
)

func (m *Model) View() string {
	if m.selectingIndex {
		return m.viewIndexSelector()
	}
	if m.viewingDoc {
		return m.viewDocDetail()
	}
	if m.timeRangeActive {
		return m.viewTimeRange()
	}
	if m.pickingTime {
		return m.viewTimePicker()
	}
	if m.buildingQuery {
		return m.viewQueryBuilder()
	}
	if m.loading && len(m.hits) == 0 {
		return "\n  Loading " + m.index + "..."
	}
	if m.err != nil {
		return "\n  " + theme.ErrorStyle.Render("Error: "+m.err.Error()) + "\n\n  " + theme.HelpDescStyle.Render("Press esc to go back")
	}

	var b strings.Builder

	// Header bar: index + time range + query + follow status
	info := theme.HelpKeyStyle.Render(m.index)
	tr := timeRanges[m.timeRangeIdx]
	info += "  " + theme.HealthYellowStyle.Render(tr.Label)
	if m.followInterval == 0 {
		info += theme.HelpDescStyle.Render(fmt.Sprintf("  %d total", m.total))
		if len(m.hits) > 0 && int64(len(m.hits)) < m.total {
			info += theme.HelpDescStyle.Render(fmt.Sprintf("  showing %d", len(m.hits)))
		}
	}
	if m.query != "" {
		info += theme.HelpDescStyle.Render("  query: ") + theme.HelpKeyStyle.Render(m.query)
	}
	if m.followInterval > 0 {
		info += "  " + theme.SuccessStyle.Render(fmt.Sprintf("▶ LIVE %s", m.followInterval))
	} else if m.index != "" {
		info += "  " + theme.HelpDescStyle.Render("⏸ PAUSED")
	}
	b.WriteString(" " + info + "\n")

	if m.editing {
		b.WriteString(" " + m.queryInput.View())
	}
	b.WriteString("\n")

	if m.pickingColumns {
		b.WriteString(m.viewColumnPicker())
	} else {
		tableView := m.table.View()
		tableView = colorizeLogLevel(tableView)
		tableView = highlightSelectedRow(tableView, m.table.Cursor(), m.hits, m.width)
		b.WriteString(tableView)
	}

	return b.String()
}

func highlightSelectedRow(view string, cursor int, hits []es.SearchHit, width int) string {
	if cursor < 0 || cursor >= len(hits) {
		return view
	}
	// The row number is cursor+1, find it in the rendered output
	rowNum := fmt.Sprintf("%d", cursor+1)
	lines := strings.Split(view, "\n")

	// Find the line containing this row number at the start (after spaces)
	// The # column shows the number, so we look for it
	selectedDone := false
	for i, line := range lines {
		if selectedDone {
			break
		}
		trimmed := strings.TrimLeft(line, " ")
		// Check if line starts with the row number followed by space
		if strings.HasPrefix(trimmed, rowNum+" ") || strings.HasPrefix(trimmed, rowNum+"\t") {
			// Pad line to full width for a wider highlight bar (use visible width, ignoring ANSI codes)
			visibleLen := lipgloss.Width(line)
			if visibleLen < width {
				line = line + strings.Repeat(" ", width-visibleLen)
			}
			lines[i] = theme.TableSelectedStyle.Render(line)
			selectedDone = true
		}
	}
	return strings.Join(lines, "\n")
}

func colorizeLogLevel(view string) string {
	// Post-process table output to colorize log levels
	view = strings.ReplaceAll(view, "ERROR", theme.ErrorStyle.Render("ERROR"))
	view = strings.ReplaceAll(view, "WARN", theme.HealthYellowStyle.Render("WARN"))
	view = strings.ReplaceAll(view, "INFO", theme.HealthGreenStyle.Render("INFO"))
	view = strings.ReplaceAll(view, "DEBUG", theme.HelpDescStyle.Render("DEBUG"))
	return view
}

func (m *Model) viewIndexSelector() string {
	var b strings.Builder
	b.WriteString("\n  " + theme.ModalTitleStyle.Render("Select Index") + "\n\n")

	if m.loading {
		b.WriteString("  Loading indices...\n")
		return b.String()
	}
	if m.err != nil {
		b.WriteString("  " + theme.ErrorStyle.Render("Error: "+m.err.Error()) + "\n")
		return b.String()
	}

	if m.indexSearching {
		b.WriteString("  " + m.indexInput.View() + "\n\n")
	} else {
		b.WriteString(theme.HelpDescStyle.Render("  Press / to filter, enter to select") + "\n\n")
	}

	maxVisible := m.height - 8
	if maxVisible < 5 {
		maxVisible = 5
	}
	start := 0
	if m.indexCursor >= maxVisible {
		start = m.indexCursor - maxVisible + 1
	}

	for i := start; i < len(m.filteredIdx) && i < start+maxVisible; i++ {
		cursor := "  "
		style := theme.HelpDescStyle
		if i == m.indexCursor {
			cursor = "> "
			style = theme.HelpKeyStyle
		}
		b.WriteString("  " + cursor + style.Render(m.filteredIdx[i]) + "\n")
	}

	if len(m.filteredIdx) == 0 {
		b.WriteString("  " + theme.HelpDescStyle.Render("(no matching indices)") + "\n")
	}

	return b.String()
}

func (m *Model) viewColumnPicker() string {
	var b strings.Builder
	b.WriteString("  " + theme.ModalTitleStyle.Render("Select Columns") + "  (space=toggle, enter=confirm)\n\n")

	maxVisible := m.height - 8
	if maxVisible < 5 {
		maxVisible = 5
	}
	start := 0
	if m.columnCursor >= maxVisible {
		start = m.columnCursor - maxVisible + 1
	}

	for i := start; i < len(m.allFields) && i < start+maxVisible; i++ {
		f := m.allFields[i]
		check := "[ ]"
		if m.columnSelected[f.Name] {
			check = "[x]"
		}
		cursor := "  "
		style := theme.HelpDescStyle
		if i == m.columnCursor {
			cursor = "> "
			style = theme.HelpKeyStyle
		}
		b.WriteString("  " + cursor + check + " " + style.Render(f.Name) + theme.HelpDescStyle.Render(" ("+f.Type+")") + "\n")
	}
	return b.String()
}

func (m *Model) viewTimePicker() string {
	var b strings.Builder
	b.WriteString("\n  " + theme.ModalTitleStyle.Render("Time Range") + "\n\n")

	for i, tr := range timeRanges {
		cursor := "  "
		style := theme.HelpDescStyle
		if i == m.timeCursor {
			cursor = theme.HelpKeyStyle.Render("> ")
			style = theme.HelpKeyStyle
		}
		marker := ""
		if i == m.timeRangeIdx {
			marker = theme.SuccessStyle.Render(" ●")
		}
		b.WriteString("  " + cursor + style.Render(tr.Label) + marker + "\n")
	}

	if m.timeCustomActive {
		b.WriteString("\n  " + m.timeCustomInput.View() + "\n")
	}

	b.WriteString("\n  " + theme.HelpDescStyle.Render("enter: select • esc: cancel") + "\n")
	return b.String()
}

func (m *Model) viewTimeRange() string {
	var b strings.Builder
	b.WriteString("\n  " + theme.ModalTitleStyle.Render("Absolute Time Range") + "\n\n")

	fromLabel := theme.HelpDescStyle.Render("  From")
	toLabel := theme.HelpDescStyle.Render("    To")
	if m.timeRangeField == 0 {
		fromLabel = theme.HelpKeyStyle.Render("▸ From")
	}
	if m.timeRangeField == 1 {
		toLabel = theme.HelpKeyStyle.Render("▸   To")
	}

	b.WriteString("  " + fromLabel + "  " + m.timeFromInput.View() + "\n")
	b.WriteString("  " + toLabel + "  " + m.timeToInput.View() + "\n")

	b.WriteString("\n  " + theme.HelpDescStyle.Render("tab: switch field • enter: apply • esc: cancel") + "\n")
	b.WriteString("  " + theme.HelpDescStyle.Render("format: 2026-03-30 13:00:00  (leave To empty = now)") + "\n")
	return b.String()
}

func (m *Model) viewDocDetail() string {
	var b strings.Builder

	docNum := m.table.Cursor() + 1
	total := len(m.hits)
	b.WriteString("\n  " + theme.ModalTitleStyle.Render(fmt.Sprintf("Document #%d / %d", docNum, total)) + "\n\n")

	b.WriteString(m.docContent)

	b.WriteString("\n  " + theme.HelpDescStyle.Render("]/j: next doc • [/k: prev doc • esc: close") + "\n")
	return b.String()
}
