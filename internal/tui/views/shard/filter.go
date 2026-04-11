package shard

import (
	"strings"

	"github.com/kienlt/es-cli/internal/es"
)

func isSystemIndex(name string) bool {
	return strings.HasPrefix(name, ".")
}

func (m *Model) countHidden() int {
	count := 0
	seen := map[string]bool{}
	for _, s := range m.shards {
		if isSystemIndex(s.Index) && !seen[s.Index] {
			seen[s.Index] = true
			count++
		}
	}
	return count
}

func (m *Model) applyFilter() {
	filtered := make([]es.Shard, 0, len(m.shards))
	query := strings.ToLower(m.filter)

	for _, s := range m.shards {
		// Hide system indices unless showAll
		if !m.showAll && isSystemIndex(s.Index) {
			continue
		}
		if query != "" {
			indexMatch := strings.Contains(strings.ToLower(s.Index), query)
			nodeMatch := strings.Contains(strings.ToLower(s.Node), query)
			if !indexMatch && !nodeMatch {
				continue
			}
		}
		filtered = append(filtered, s)
	}
	m.filtered = filtered
}
