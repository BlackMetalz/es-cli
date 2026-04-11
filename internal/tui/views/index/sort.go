package index

import (
	"sort"

	"github.com/kienlt/es-cli/internal/es"
)

func (m *Model) sortIndices() {
	asc := m.sortAsc
	switch m.sortField {
	case SortByName:
		sort.Slice(m.indices, func(i, j int) bool {
			cmp := naturalCompare(m.indices[i].Name, m.indices[j].Name)
			if asc {
				return cmp < 0
			}
			return cmp > 0
		})
	case SortByHealth:
		sort.Slice(m.indices, func(i, j int) bool {
			if asc {
				return healthRank(m.indices[i].Health) < healthRank(m.indices[j].Health)
			}
			return healthRank(m.indices[i].Health) > healthRank(m.indices[j].Health)
		})
	case SortBySize:
		sort.Slice(m.indices, func(i, j int) bool {
			sizeI := es.ParseSizeToBytes(m.indices[i].StoreSize)
			sizeJ := es.ParseSizeToBytes(m.indices[j].StoreSize)
			if asc {
				return sizeI < sizeJ
			}
			return sizeI > sizeJ
		})
	case SortByDocs:
		sort.Slice(m.indices, func(i, j int) bool {
			docsI := es.ParseSizeToBytes(m.indices[i].DocsCount)
			docsJ := es.ParseSizeToBytes(m.indices[j].DocsCount)
			if asc {
				return docsI < docsJ
			}
			return docsI > docsJ
		})
	}
}

// naturalCompare compares two strings with natural number ordering.
// "demo-2" < "demo-10" instead of lexicographic "demo-10" < "demo-2".
func naturalCompare(a, b string) int {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		ca, cb := a[i], b[j]
		if isDigit(ca) && isDigit(cb) {
			na, ni := extractNumber(a, i)
			nb, nj := extractNumber(b, j)
			if na != nb {
				if na < nb {
					return -1
				}
				return 1
			}
			i, j = ni, nj
		} else {
			if ca != cb {
				if ca < cb {
					return -1
				}
				return 1
			}
			i++
			j++
		}
	}
	return len(a) - len(b)
}

func extractNumber(s string, i int) (int64, int) {
	var n int64
	for i < len(s) && isDigit(s[i]) {
		n = n*10 + int64(s[i]-'0')
		i++
	}
	return n, i
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func healthRank(health string) int {
	switch health {
	case "green":
		return 0
	case "yellow":
		return 1
	case "red":
		return 2
	default:
		return 3
	}
}
