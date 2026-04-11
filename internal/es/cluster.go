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
