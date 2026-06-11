package threadpool

import (
	"sort"
	"strings"
)

type SortField int

const (
	SortByName SortField = iota
	SortByNode
	SortByActive
	SortByQueue
	SortByRejected
)

func (m *Model) sortPools() {
	asc := m.sortAsc
	sort.SliceStable(m.pools, func(i, j int) bool {
		a, b := m.pools[i], m.pools[j]
		switch m.sortField {
		case SortByNode:
			cmp := strings.Compare(strings.ToLower(a.Node), strings.ToLower(b.Node))
			if cmp == 0 {
				return strings.ToLower(a.Name) < strings.ToLower(b.Name)
			}
			if asc {
				return cmp < 0
			}
			return cmp > 0
		case SortByActive:
			if asc {
				return a.Active < b.Active
			}
			return a.Active > b.Active
		case SortByQueue:
			if asc {
				return a.Queue < b.Queue
			}
			return a.Queue > b.Queue
		case SortByRejected:
			if asc {
				return a.Rejected < b.Rejected
			}
			return a.Rejected > b.Rejected
		default: // SortByName: group same pool names together across nodes
			cmp := strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
			if cmp == 0 {
				return strings.ToLower(a.Node) < strings.ToLower(b.Node)
			}
			if asc {
				return cmp < 0
			}
			return cmp > 0
		}
	})
}

func (m *Model) toggleSort(field SortField) {
	if m.sortField == field {
		m.sortAsc = !m.sortAsc
	} else {
		m.sortField = field
		m.sortAsc = true
	}
	m.sortPools()
	m.applyFilter()
	m.updateTable()
}
