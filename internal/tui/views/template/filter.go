package template

import (
	"strings"

	"github.com/kienlt/es-cli/internal/es"
)

func isHiddenTemplate(t es.IndexTemplate) bool {
	return strings.HasPrefix(t.Name, ".") || t.Managed
}

func (m *Model) countHidden() int {
	count := 0
	for _, t := range m.templates {
		if isHiddenTemplate(t) {
			count++
		}
	}
	return count
}

func (m *Model) applyFilter() {
	filtered := make([]es.IndexTemplate, 0, len(m.templates))
	query := strings.ToLower(m.filter)

	for _, t := range m.templates {
		if !m.showAll && isHiddenTemplate(t) {
			continue
		}
		if query != "" {
			nameLower := strings.ToLower(t.Name)
			patternsLower := strings.ToLower(t.IndexPatterns)
			if !strings.Contains(nameLower, query) && !strings.Contains(patternsLower, query) {
				continue
			}
		}
		filtered = append(filtered, t)
	}
	m.filtered = filtered
}
