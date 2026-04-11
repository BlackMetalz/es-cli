package index

import (
	"strings"

	"github.com/kienlt/es-cli/internal/es"
)

func isSystemIndex(name string) bool {
	return strings.HasPrefix(name, ".")
}

func (m *Model) countHidden() int {
	count := 0
	for _, idx := range m.indices {
		if isSystemIndex(idx.Name) {
			count++
		}
	}
	return count
}

func (m *Model) applyFilter() {
	filtered := make([]es.Index, 0, len(m.indices))
	query := strings.ToLower(m.filter)

	for _, idx := range m.indices {
		// Hide system indices unless showAll
		if !m.showAll && isSystemIndex(idx.Name) {
			continue
		}
		// Apply search filter
		if query != "" && !strings.Contains(strings.ToLower(idx.Name), query) {
			continue
		}
		filtered = append(filtered, idx)
	}
	m.filtered = filtered
}
