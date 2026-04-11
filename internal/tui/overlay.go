package tui

// placeOverlay places an overlay string on top of a background string at the given position.
func placeOverlay(x, y int, overlay, background string) string {
	bgLines := splitLines(background)
	fgLines := splitLines(overlay)

	for i, fgLine := range fgLines {
		bgIdx := y + i
		if bgIdx < 0 || bgIdx >= len(bgLines) {
			continue
		}

		bgLine := bgLines[bgIdx]
		bgRunes := []rune(bgLine)

		for len(bgRunes) < x+len([]rune(fgLine)) {
			bgRunes = append(bgRunes, ' ')
		}

		fgRunes := []rune(fgLine)
		for j, r := range fgRunes {
			pos := x + j
			if pos >= 0 && pos < len(bgRunes) {
				bgRunes[pos] = r
			}
		}

		bgLines[bgIdx] = string(bgRunes)
	}

	result := ""
	for i, line := range bgLines {
		if i > 0 {
			result += "\n"
		}
		result += line
	}
	return result
}

func splitLines(s string) []string {
	lines := []string{}
	current := ""
	for _, r := range s {
		if r == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	lines = append(lines, current)
	return lines
}
