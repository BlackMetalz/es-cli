package ilm

import (
	"sort"
	"strings"
)

func (m *Model) sortPolicies() {
	asc := m.sortAsc
	sort.Slice(m.policies, func(i, j int) bool {
		cmp := strings.Compare(
			strings.ToLower(m.policies[i].Name),
			strings.ToLower(m.policies[j].Name),
		)
		if asc {
			return cmp < 0
		}
		return cmp > 0
	})
}
