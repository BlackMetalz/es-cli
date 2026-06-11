package threadpool

import (
	"strings"

	"github.com/kienlt/es-cli/internal/es"
)

func (m *Model) applyFilter() {
	filtered := make([]es.ThreadPool, 0, len(m.pools))
	query := strings.ToLower(m.filter)

	for _, tp := range m.pools {
		if !m.showAll && tp.Type == "direct" {
			continue
		}
		if query != "" {
			if !strings.Contains(strings.ToLower(tp.Name), query) &&
				!strings.Contains(strings.ToLower(tp.Node), query) {
				continue
			}
		}
		filtered = append(filtered, tp)
	}
	m.filtered = filtered
}

func (m *Model) countHidden() int {
	count := 0
	for _, tp := range m.pools {
		if tp.Type == "direct" {
			count++
		}
	}
	return count
}
