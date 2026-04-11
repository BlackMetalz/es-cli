package node

import (
	"sort"
	"strconv"
	"strings"
)

type SortField int

const (
	SortByName SortField = iota
	SortByCPU
	SortByHeap
	SortByRAM
	SortByDisk
)

func (m *Model) sortNodes() {
	asc := m.sortAsc
	switch m.sortField {
	case SortByName:
		sort.Slice(m.nodes, func(i, j int) bool {
			if asc {
				return strings.ToLower(m.nodes[i].Name) < strings.ToLower(m.nodes[j].Name)
			}
			return strings.ToLower(m.nodes[i].Name) > strings.ToLower(m.nodes[j].Name)
		})
	case SortByCPU:
		sort.Slice(m.nodes, func(i, j int) bool {
			return compareFloat(m.nodes[i].CPU, m.nodes[j].CPU, asc)
		})
	case SortByHeap:
		sort.Slice(m.nodes, func(i, j int) bool {
			return compareFloat(m.nodes[i].HeapPercent, m.nodes[j].HeapPercent, asc)
		})
	case SortByRAM:
		sort.Slice(m.nodes, func(i, j int) bool {
			return compareFloat(m.nodes[i].RAMPercent, m.nodes[j].RAMPercent, asc)
		})
	case SortByDisk:
		sort.Slice(m.nodes, func(i, j int) bool {
			return compareFloat(m.nodes[i].DiskUsedPercent, m.nodes[j].DiskUsedPercent, asc)
		})
	}
}

func compareFloat(a, b string, asc bool) bool {
	va := parsePercent(a)
	vb := parsePercent(b)
	if asc {
		return va < vb
	}
	return va > vb
}

func parsePercent(s string) float64 {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, "%")
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}

// sortFieldLabel returns the column name for the current sort field.
func sortFieldLabel(field SortField) string {
	switch field {
	case SortByName:
		return "name"
	case SortByCPU:
		return "cpu"
	case SortByHeap:
		return "heap%"
	case SortByRAM:
		return "ram%"
	case SortByDisk:
		return "disk%"
	default:
		return "name"
	}
}
