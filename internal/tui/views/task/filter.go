package task

import (
	"strings"

	"github.com/kienlt/es-cli/internal/es"
)

type CategoryFilter int

const (
	CategoryAll      CategoryFilter = iota
	CategoryRead                    // indices:data/read/*
	CategoryWrite                   // indices:data/write/*
	CategoryAdmin                   // *admin* (excluding snapshot)
	CategorySnapshot                // *snapshot*
)

var categoryNames = []string{"all", "read", "write", "admin", "snapshot"}

func (c CategoryFilter) String() string {
	if int(c) < len(categoryNames) {
		return categoryNames[c]
	}
	return "all"
}

func (c CategoryFilter) Matches(action string) bool {
	switch c {
	case CategoryRead:
		return strings.Contains(action, "data/read")
	case CategoryWrite:
		return strings.Contains(action, "data/write")
	case CategoryAdmin:
		return strings.Contains(action, "admin") && !strings.Contains(action, "snapshot")
	case CategorySnapshot:
		return strings.Contains(action, "snapshot")
	default:
		return true
	}
}

func (m *Model) nextCategory() {
	m.category = CategoryFilter((int(m.category) + 1) % len(categoryNames))
	m.applyFilter()
	m.updateTable()
}

func (m *Model) applyFilter() {
	query := strings.ToLower(m.filter)
	var intermediate []es.Task

	for _, t := range m.tasks {
		if !m.category.Matches(t.Action) {
			continue
		}
		if query != "" {
			if !strings.Contains(strings.ToLower(t.Action), query) &&
				!strings.Contains(strings.ToLower(t.NodeName), query) &&
				!strings.Contains(strings.ToLower(t.Description), query) {
				continue
			}
		}
		intermediate = append(intermediate, t)
	}

	m.filteredCount = len(intermediate)

	if !m.showAll && len(intermediate) > topN {
		m.filtered = intermediate[:topN]
	} else {
		m.filtered = intermediate
	}
}
