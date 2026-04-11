package es

import (
	"encoding/json"
	"fmt"
)

type Node struct {
	Name            string
	IP              string
	HeapPercent     string
	RAMPercent      string
	CPU             string
	Load1m          string
	Load5m          string
	Load15m         string
	NodeRole        string
	Master          string
	DiskUsedPercent string
}

func (c *Client) ListNodes() ([]Node, error) {
	data, err := c.Get("/_cat/nodes?format=json&h=name,ip,heap.percent,ram.percent,cpu,load_1m,load_5m,load_15m,node.role,master,disk.used_percent")
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var raw []map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse nodes response: %w", err)
	}

	nodes := make([]Node, len(raw))
	for i, r := range raw {
		nodes[i] = Node{
			Name:            JsonStr(r["name"]),
			IP:              JsonStr(r["ip"]),
			HeapPercent:     JsonStr(r["heap.percent"]),
			RAMPercent:      JsonStr(r["ram.percent"]),
			CPU:             JsonStr(r["cpu"]),
			Load1m:          JsonStr(r["load_1m"]),
			Load5m:          JsonStr(r["load_5m"]),
			Load15m:         JsonStr(r["load_15m"]),
			NodeRole:        JsonStr(r["node.role"]),
			Master:          JsonStr(r["master"]),
			DiskUsedPercent: JsonStr(r["disk.used_percent"]),
		}
	}

	return nodes, nil
}
