package node

import (
	"strings"

	"github.com/kienlt/es-cli/internal/es"
)

func (m *Model) applyFilter() {
	filtered := make([]es.Node, 0, len(m.nodes))
	query := strings.ToLower(m.filter)

	for _, n := range m.nodes {
		if query != "" && !strings.Contains(strings.ToLower(n.Name), query) {
			continue
		}
		filtered = append(filtered, n)
	}
	m.filtered = filtered
}
