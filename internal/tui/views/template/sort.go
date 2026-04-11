package template

import (
	"sort"
	"strings"
)

func (m *Model) sortTemplates() {
	asc := m.sortAsc
	switch m.sortField {
	case SortByName:
		sort.Slice(m.templates, func(i, j int) bool {
			cmp := strings.Compare(
				strings.ToLower(m.templates[i].Name),
				strings.ToLower(m.templates[j].Name),
			)
			if asc {
				return cmp < 0
			}
			return cmp > 0
		})
	}
}
