package es

import (
	"encoding/json"
	"fmt"
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

	return d, nil
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
