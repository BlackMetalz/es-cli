package es

import (
	"encoding/json"
	"fmt"
)

type ThreadPool struct {
	Node      string
	Name      string
	Type      string
	Active    int
	Size      int // configured max threads (0 for scaling pools)
	Queue     int
	QueueSize int // -1 for unbounded
	Rejected  int
	Largest   int
	Completed int
}

func (c *Client) ListThreadPools() ([]ThreadPool, error) {
	data, err := c.Get("/_cat/thread_pool?format=json&h=node_name,name,type,active,size,queue,queue_size,rejected,largest,completed")
	if err != nil {
		return nil, fmt.Errorf("failed to list thread pools: %w", err)
	}

	var raw []map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse thread pools response: %w", err)
	}

	pools := make([]ThreadPool, len(raw))
	for i, r := range raw {
		queueSize := -1
		qsStr := JsonStr(r["queue_size"])
		if qsStr != "unlimited" && qsStr != "" {
			if v := jsonInt(r["queue_size"]); v >= 0 {
				queueSize = v
			}
		}

		pools[i] = ThreadPool{
			Node:      JsonStr(r["node_name"]),
			Name:      JsonStr(r["name"]),
			Type:      JsonStr(r["type"]),
			Active:    jsonInt(r["active"]),
			Size:      jsonInt(r["size"]),
			Queue:     jsonInt(r["queue"]),
			QueueSize: queueSize,
			Rejected:  jsonInt(r["rejected"]),
			Largest:   jsonInt(r["largest"]),
			Completed: jsonInt(r["completed"]),
		}
	}

	return pools, nil
}
