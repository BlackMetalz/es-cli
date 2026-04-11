package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/kienlt/es-cli/internal/tui/theme"
	"github.com/kienlt/es-cli/internal/tui/views"
)

func renderHelpFullScreen(groups []views.HelpGroup, width, height int) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(theme.ColorCyan).
		Bold(true)

	groupTitleStyle := lipgloss.NewStyle().
		Foreground(theme.ColorYellow).
		Bold(true)

	type column struct {
		lines []string
		width int
	}

	var columns []column
	for _, g := range groups {
		var lines []string
		maxWidth := len(g.Title)

		lines = append(lines, groupTitleStyle.Render(g.Title))

		for _, b := range g.Bindings {
			h := b.Help()
			keyStr := theme.HelpKeyStyle.Render(fmt.Sprintf("<%s>", h.Key))
			descStr := theme.HelpDescStyle.Render("  " + h.Desc)
			line := keyStr + descStr
			lines = append(lines, line)

			w := lipgloss.Width(line)
			if w > maxWidth {
				maxWidth = w
			}
		}

		columns = append(columns, column{lines: lines, width: maxWidth + 4})
	}

	maxRows := 0
	for _, col := range columns {
		if len(col.lines) > maxRows {
			maxRows = len(col.lines)
		}
	}

	var bodyLines []string
	for row := 0; row < maxRows; row++ {
		var rowParts []string
		for _, col := range columns {
			cell := ""
			if row < len(col.lines) {
				cell = col.lines[row]
			}
			visible := lipgloss.Width(cell)
			pad := col.width - visible
			if pad < 0 {
				pad = 0
			}
			rowParts = append(rowParts, cell+strings.Repeat(" ", pad))
		}
		bodyLines = append(bodyLines, " "+strings.Join(rowParts, ""))
	}

	// Title bar
	title := titleStyle.Render("Help")
	titlePad := (width - lipgloss.Width(title)) / 2
	if titlePad < 0 {
		titlePad = 0
	}
	titleLine := strings.Repeat("─", titlePad) + " " + title + " " + strings.Repeat("─", titlePad)
	titleLine = theme.SeparatorStyle.Render(titleLine)

	footer := theme.HelpDescStyle.Render("Press ") +
		theme.HelpKeyStyle.Render("Esc") +
		theme.HelpDescStyle.Render(" or ") +
		theme.HelpKeyStyle.Render("?") +
		theme.HelpDescStyle.Render(" to close")

	var out []string
	out = append(out, titleLine)
	out = append(out, bodyLines...)

	used := len(out) + 1
	for i := 0; i < height-used; i++ {
		out = append(out, "")
	}
	out = append(out, " "+footer)

	return strings.Join(out, "\n")
}
