package es

import (
	"encoding/json"
	"fmt"
)

// GetAllocationSetting returns the current cluster routing allocation setting.
// Returns "" if default (all), otherwise "primaries" or "none".
func (c *Client) GetAllocationSetting() (string, error) {
	data, err := c.Get("/_cluster/settings?flat_settings=true")
	if err != nil {
		return "", fmt.Errorf("failed to get cluster settings: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return "", fmt.Errorf("failed to parse cluster settings: %w", err)
	}

	// Check transient first, then persistent
	for _, section := range []string{"transient", "persistent"} {
		if s, ok := raw[section].(map[string]interface{}); ok {
			if val, ok := s["cluster.routing.allocation.enable"].(string); ok && val != "" && val != "all" {
				return val, nil
			}
		}
	}

	return "", nil
}

// SetAllocationSetting sets the cluster routing allocation setting.
// Pass "" to reset to default (all).
func (c *Client) SetAllocationSetting(value string) error {
	var body string
	if value == "" {
		body = `{"transient":{"cluster.routing.allocation.enable":null}}`
	} else {
		body = fmt.Sprintf(`{"transient":{"cluster.routing.allocation.enable":"%s"}}`, value)
	}

	_, err := c.Put("/_cluster/settings", body)
	if err != nil {
		return fmt.Errorf("failed to set allocation setting: %w", err)
	}
	return nil
}

// RetryFailedAllocation calls POST /_cluster/reroute?retry_failed=true to retry
// shard allocations that previously failed past the max-retries limit.
func (c *Client) RetryFailedAllocation() error {
	_, err := c.Post("/_cluster/reroute?retry_failed=true", "")
	if err != nil {
		return fmt.Errorf("failed to retry shard allocation: %w", err)
	}
	return nil
}

// ClearCache calls POST /_cache/clear to drop field data, query, and request
// caches cluster-wide, freeing the heap they occupy. Caches rebuild on next
// access, so this is non-destructive but causes a temporary cache-miss cost.
func (c *Client) ClearCache() error {
	_, err := c.Post("/_cache/clear", "")
	if err != nil {
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	return nil
}

// ClearAllScrolls calls DELETE /_search/scroll/_all to release every open
// scroll context cluster-wide immediately, instead of waiting for its
// keep_alive to expire. Open scroll contexts pin heap and segment files and
// count against search.max_open_scroll_context, so a client that leaks
// scrolls (opens without clearing) can starve the cluster of new ones.
func (c *Client) ClearAllScrolls() error {
	_, err := c.Delete("/_search/scroll/_all")
	if err != nil {
		return fmt.Errorf("failed to clear scrolls: %w", err)
	}
	return nil
}
