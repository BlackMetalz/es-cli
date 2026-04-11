package shard

import (
	"sort"
	"strconv"
	"strings"

	"github.com/kienlt/es-cli/internal/es"
)

func (m *Model) sortShards() {
	asc := m.sortAsc
	switch m.sortField {
	case SortByIndex:
		sort.Slice(m.shards, func(i, j int) bool {
			cmp := strings.Compare(
				strings.ToLower(m.shards[i].Index),
				strings.ToLower(m.shards[j].Index),
			)
			if cmp == 0 {
				// Secondary sort by shard number
				si, _ := strconv.Atoi(m.shards[i].ShardN)
				sj, _ := strconv.Atoi(m.shards[j].ShardN)
				return si < sj
			}
			if asc {
				return cmp < 0
			}
			return cmp > 0
		})
	case SortByShard:
		sort.Slice(m.shards, func(i, j int) bool {
			si, _ := strconv.Atoi(m.shards[i].ShardN)
			sj, _ := strconv.Atoi(m.shards[j].ShardN)
			if asc {
				return si < sj
			}
			return si > sj
		})
	case SortByState:
		sort.Slice(m.shards, func(i, j int) bool {
			cmp := strings.Compare(
				strings.ToLower(m.shards[i].State),
				strings.ToLower(m.shards[j].State),
			)
			if asc {
				return cmp < 0
			}
			return cmp > 0
		})
	case SortByNode:
		sort.Slice(m.shards, func(i, j int) bool {
			cmp := strings.Compare(
				strings.ToLower(m.shards[i].Node),
				strings.ToLower(m.shards[j].Node),
			)
			if asc {
				return cmp < 0
			}
			return cmp > 0
		})
	case SortByDocs:
		sort.Slice(m.shards, func(i, j int) bool {
			di, _ := strconv.Atoi(m.shards[i].Docs)
			dj, _ := strconv.Atoi(m.shards[j].Docs)
			if asc {
				return di < dj
			}
			return di > dj
		})
	case SortByStore:
		sort.Slice(m.shards, func(i, j int) bool {
			si := es.ParseSizeToBytes(m.shards[i].Store)
			sj := es.ParseSizeToBytes(m.shards[j].Store)
			if asc {
				return si < sj
			}
			return si > sj
		})
	}
}
