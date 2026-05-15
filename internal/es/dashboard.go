package es

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"time"
)

type DashboardData struct {
	// Overview
	Health            string
	HealthDescription string
	Version           string
	Uptime            time.Duration
	License           string

	// Nodes
	NodeCount      int
	DiskAvailBytes float64
	DiskTotalBytes float64
	HeapUsedBytes  float64
	HeapMaxBytes   float64

	// Indices
	IndexCount    int
	DocCount      int64
	DiskUsage     string
	PrimaryShards int
	ReplicaShards int

	// Index pattern breakdown
	PatternStats []IndexPatternStat
}

// IndexPatternStat holds aggregated stats for a group of similarly-named indices.
type IndexPatternStat struct {
	Pattern    string // e.g. "demo-*" or "app-logs"
	IndexCount int
	Shards     int
	DiskBytes  int64
}

// DiskAvailPercent returns disk available as a percentage.
func (d *DashboardData) DiskAvailPercent() float64 {
	if d.DiskTotalBytes == 0 {
		return 0
	}
	return (d.DiskAvailBytes / d.DiskTotalBytes) * 100
}

// HeapUsedPercent returns JVM heap used as a percentage.
func (d *DashboardData) HeapUsedPercent() float64 {
	if d.HeapMaxBytes == 0 {
		return 0
	}
	return (d.HeapUsedBytes / d.HeapMaxBytes) * 100
}

var healthDescriptions = map[string]string{
	"green":  "All shards assigned",
	"yellow": "Missing replica shards",
	"red":    "Missing primary shards",
}

func (c *Client) GetDashboardData() (*DashboardData, error) {
	d := &DashboardData{}

	// 1. GET / — version
	rootData, err := c.Get("/")
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster root: %w", err)
	}
	var root map[string]interface{}
	if err := json.Unmarshal(rootData, &root); err == nil {
		if ver, ok := root["version"].(map[string]interface{}); ok {
			d.Version = JsonStr(ver["number"])
		}
	}

	// 2. GET /_cluster/stats — indices + nodes aggregated stats
	statsData, err := c.Get("/_cluster/stats")
	if err != nil {
		return nil, fmt.Errorf("failed to get cluster stats: %w", err)
	}
	var stats map[string]interface{}
	if err := json.Unmarshal(statsData, &stats); err == nil {
		d.Health = JsonStr(stats["status"])
		d.HealthDescription = healthDescriptions[d.Health]

		if indices, ok := stats["indices"].(map[string]interface{}); ok {
			d.IndexCount = jsonInt(indices["count"])
			if docs, ok := indices["docs"].(map[string]interface{}); ok {
				d.DocCount = jsonInt64(docs["count"])
			}
			if store, ok := indices["store"].(map[string]interface{}); ok {
				d.DiskUsage = FormatBytes(fmt.Sprintf("%.0f", jsonFloat(store["size_in_bytes"])))
			}
			if shards, ok := indices["shards"].(map[string]interface{}); ok {
				d.PrimaryShards = jsonInt(shards["primaries"])
				total := jsonInt(shards["total"])
				d.ReplicaShards = total - d.PrimaryShards
			}
		}

		if nodes, ok := stats["nodes"].(map[string]interface{}); ok {
			if count, ok := nodes["count"].(map[string]interface{}); ok {
				d.NodeCount = jsonInt(count["total"])
			}
			if jvm, ok := nodes["jvm"].(map[string]interface{}); ok {
				if mem, ok := jvm["mem"].(map[string]interface{}); ok {
					d.HeapUsedBytes = jsonFloat(mem["heap_used_in_bytes"])
					d.HeapMaxBytes = jsonFloat(mem["heap_max_in_bytes"])
				}
			}
			if fs, ok := nodes["fs"].(map[string]interface{}); ok {
				d.DiskTotalBytes = jsonFloat(fs["total_in_bytes"])
				d.DiskAvailBytes = jsonFloat(fs["available_in_bytes"])
			}
		}
	}

	// 3. GET /_license — license type
	licenseData, err := c.Get("/_license")
	if err == nil {
		var lic map[string]interface{}
		if json.Unmarshal(licenseData, &lic) == nil {
			if license, ok := lic["license"].(map[string]interface{}); ok {
				d.License = JsonStr(license["type"])
			}
		}
	}

	// 4. GET /_nodes/stats/jvm — uptime from longest-running node
	jvmData, err := c.Get("/_nodes/stats/jvm")
	if err == nil {
		var jvmStats map[string]interface{}
		if json.Unmarshal(jvmData, &jvmStats) == nil {
			if nodes, ok := jvmStats["nodes"].(map[string]interface{}); ok {
				var maxUptime int64
				for _, nodeData := range nodes {
					if node, ok := nodeData.(map[string]interface{}); ok {
						if jvm, ok := node["jvm"].(map[string]interface{}); ok {
							uptimeMs := int64(jsonFloat(jvm["uptime_in_millis"]))
							if uptimeMs > maxUptime {
								maxUptime = uptimeMs
							}
						}
					}
				}
				if maxUptime > 0 {
					d.Uptime = time.Duration(maxUptime) * time.Millisecond
				}
			}
		}
	}

	// 5. Index pattern stats (includes hidden indices)
	if ps, err := c.GetIndexPatternStats(); err == nil {
		d.PatternStats = ps
	}

	return d, nil
}

// trailingNumRe matches a numeric suffix preceded by a separator (-, _, .)
// e.g. "demo-1" → ["demo-1", "demo", "-", "1"]
var trailingNumRe = regexp.MustCompile(`^(.*?)([-._])(\d[\d.]*)$`)

// GetIndexPatternStats fetches all indices (including hidden) and groups them by
// detected pattern. Indices like demo-1, demo-2 are collapsed to demo-*.
func (c *Client) GetIndexPatternStats() ([]IndexPatternStat, error) {
	data, err := c.Get("/_cat/indices?format=json&h=index,pri,rep,store.size&bytes=b&expand_wildcards=all")
	if err != nil {
		return nil, err
	}
	var rows []map[string]interface{}
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, err
	}

	type entry struct {
		name   string
		shards int
		bytes  int64
	}

	type group struct {
		pattern string
		items   []entry
	}

	groupMap := map[string]*group{}
	var keyOrder []string

	for _, row := range rows {
		name := JsonStr(row["index"])
		pri := jsonInt(row["pri"])
		rep := jsonInt(row["rep"])
		shards := pri * (1 + rep)

		var diskBytes int64
		if b, err2 := strconv.ParseInt(JsonStr(row["store.size"]), 10, 64); err2 == nil {
			diskBytes = b
		}

		e := entry{name: name, shards: shards, bytes: diskBytes}

		// Derive grouping key: "demo-" for "demo-1", "demo-2"; full name for singles
		var key string
		if m := trailingNumRe.FindStringSubmatch(name); m != nil {
			key = m[1] + m[2] // base + separator, e.g. "demo-"
		} else {
			key = name
		}

		if _, exists := groupMap[key]; !exists {
			groupMap[key] = &group{}
			keyOrder = append(keyOrder, key)
		}
		groupMap[key].items = append(groupMap[key].items, e)
	}

	stats := make([]IndexPatternStat, 0, len(groupMap))
	for _, key := range keyOrder {
		g := groupMap[key]
		var totalShards int
		var totalBytes int64
		for _, e := range g.items {
			totalShards += e.shards
			totalBytes += e.bytes
		}
		pattern := key
		if len(g.items) > 1 {
			pattern = key + "*"
		} else {
			pattern = g.items[0].name
		}
		stats = append(stats, IndexPatternStat{
			Pattern:    pattern,
			IndexCount: len(g.items),
			Shards:     totalShards,
			DiskBytes:  totalBytes,
		})
	}

	// Sort by disk usage descending so the heaviest patterns appear first
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].DiskBytes > stats[j].DiskBytes
	})

	return stats, nil
}

func jsonInt(v interface{}) int {
	switch n := v.(type) {
	case float64:
		return int(n)
	case string:
		i, _ := strconv.Atoi(n)
		return i
	}
	return 0
}

func jsonInt64(v interface{}) int64 {
	switch n := v.(type) {
	case float64:
		return int64(n)
	case string:
		i, _ := strconv.ParseInt(n, 10, 64)
		return i
	}
	return 0
}

func jsonFloat(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case string:
		f, _ := strconv.ParseFloat(n, 64)
		return f
	}
	return 0
}
