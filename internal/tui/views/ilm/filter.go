package ilm

import (
	"strings"

	"github.com/kienlt/es-cli/internal/es"
)

func isHiddenPolicy(p es.ILMPolicy) bool {
	return strings.HasPrefix(p.Name, ".") || p.Managed
}

func (m *Model) countHidden() int {
	count := 0
	for _, p := range m.policies {
		if isHiddenPolicy(p) {
			count++
		}
	}
	return count
}

func (m *Model) applyFilter() {
	filtered := make([]es.ILMPolicy, 0, len(m.policies))
	query := strings.ToLower(m.filter)

	for _, p := range m.policies {
		if !m.showAll && isHiddenPolicy(p) {
			continue
		}
		if query != "" && !strings.Contains(strings.ToLower(p.Name), query) {
			continue
		}
		filtered = append(filtered, p)
	}
	m.filtered = filtered
}
