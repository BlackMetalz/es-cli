package es

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type Index struct {
	Health    string
	Status    string
	Name      string
	Pri       string
	Rep       string
	DocsCount string
	StoreSize string
}

func (c *Client) ListIndices() ([]Index, error) {
	data, err := c.Get("/_cat/indices?format=json&h=health,status,index,pri,rep,docs.count,pri.store.size")
	if err != nil {
		return nil, fmt.Errorf("failed to list indices: %w", err)
	}

	var raw []map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse indices response: %w", err)
	}

	indices := make([]Index, len(raw))
	for i, r := range raw {
		indices[i] = Index{
			Health:    jsonStr(r["health"]),
			Status:    jsonStr(r["status"]),
			Name:      jsonStr(r["index"]),
			Pri:       jsonStr(r["pri"]),
			Rep:       jsonStr(r["rep"]),
			DocsCount: jsonStr(r["docs.count"]),
			StoreSize: jsonStr(r["pri.store.size"]),
		}
	}

	return indices, nil
}

func jsonStr(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

func (c *Client) CreateIndex(name string, shards, replicas int) error {
	body := fmt.Sprintf(`{"settings":{"number_of_shards":%d,"number_of_replicas":%d}}`, shards, replicas)
	_, err := c.Put("/"+name, body)
	if err != nil {
		return fmt.Errorf("failed to create index %s: %w", name, err)
	}
	return nil
}

func (c *Client) CloseIndex(name string) error {
	_, err := c.Post("/"+name+"/_close", "")
	if err != nil {
		return fmt.Errorf("failed to close index %s: %w", name, err)
	}
	return nil
}

func (c *Client) OpenIndex(name string) error {
	_, err := c.Post("/"+name+"/_open", "")
	if err != nil {
		return fmt.Errorf("failed to open index %s: %w", name, err)
	}
	return nil
}

func (c *Client) DeleteIndex(name string) error {
	_, err := c.Delete("/" + name)
	if err != nil {
		return fmt.Errorf("failed to delete index %s: %w", name, err)
	}
	return nil
}

// FormatBytes converts a byte count string to human-readable format (e.g., "50.0mb").
func FormatBytes(bytesStr string) string {
	bytesStr = strings.TrimSpace(bytesStr)
	if bytesStr == "" || bytesStr == "-" {
		return "-"
	}
	b, err := strconv.ParseFloat(bytesStr, 64)
	if err != nil {
		return bytesStr
	}
	if b == 0 {
		return "0b"
	}

	units := []string{"b", "kb", "mb", "gb", "tb"}
	i := 0
	for b >= 1024 && i < len(units)-1 {
		b /= 1024
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%.0fb", b)
	}
	return fmt.Sprintf("%.1f%s", b, units[i])
}

// ParseSizeToBytes converts human-readable size strings (e.g., "4.5kb", "1.2gb") to bytes.
func ParseSizeToBytes(size string) int64 {
	size = strings.TrimSpace(strings.ToLower(size))
	if size == "" || size == "-" {
		return 0
	}

	type unit struct {
		suffix string
		mult   float64
	}
	// Order matters: check longer suffixes first
	units := []unit{
		{"tb", 1024 * 1024 * 1024 * 1024},
		{"gb", 1024 * 1024 * 1024},
		{"mb", 1024 * 1024},
		{"kb", 1024},
		{"b", 1},
	}

	for _, u := range units {
		if strings.HasSuffix(size, u.suffix) {
			numStr := strings.TrimSuffix(size, u.suffix)
			num, err := strconv.ParseFloat(numStr, 64)
			if err != nil {
				return 0
			}
			return int64(math.Round(num * u.mult))
		}
	}

	// Try parsing as plain number (bytes)
	num, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		return 0
	}
	return num
}
