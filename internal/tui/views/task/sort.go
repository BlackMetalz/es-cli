package task

import (
	"sort"
	"strings"
)

type SortField int

const (
	SortByDuration SortField = iota
	SortByAction
	SortByNode
)

func (m *Model) sortTasks() {
	asc := m.sortAsc
	sort.SliceStable(m.tasks, func(i, j int) bool {
		a, b := m.tasks[i], m.tasks[j]
		switch m.sortField {
		case SortByAction:
			cmp := strings.Compare(strings.ToLower(a.Action), strings.ToLower(b.Action))
			if cmp == 0 {
				return a.RunningTimeNanos > b.RunningTimeNanos
			}
			if asc {
				return cmp < 0
			}
			return cmp > 0
		case SortByNode:
			cmp := strings.Compare(strings.ToLower(a.NodeName), strings.ToLower(b.NodeName))
			if cmp == 0 {
				return a.RunningTimeNanos > b.RunningTimeNanos
			}
			if asc {
				return cmp < 0
			}
			return cmp > 0
		default: // SortByDuration
			if asc {
				return a.RunningTimeNanos < b.RunningTimeNanos
			}
			return a.RunningTimeNanos > b.RunningTimeNanos
		}
	})
}

func (m *Model) toggleSort(field SortField) {
	if m.sortField == field {
		m.sortAsc = !m.sortAsc
	} else {
		m.sortField = field
		m.sortAsc = field != SortByDuration
	}
	m.sortTasks()
	m.applyFilter()
	m.updateTable()
}
