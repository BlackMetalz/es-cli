package es

import (
	"encoding/json"
	"fmt"
)

type IndexDetail struct {
	Settings json.RawMessage
	Mappings json.RawMessage
	Aliases  json.RawMessage
}

func (c *Client) GetIndexDetail(name string) (*IndexDetail, error) {
	detail := &IndexDetail{}

	settings, err := c.Get(fmt.Sprintf("/%s/_settings", name))
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}
	detail.Settings = settings

	mappings, err := c.Get(fmt.Sprintf("/%s/_mapping", name))
	if err != nil {
		return nil, fmt.Errorf("failed to get mappings: %w", err)
	}
	detail.Mappings = mappings

	aliases, err := c.Get(fmt.Sprintf("/%s/_alias", name))
	if err != nil {
		return nil, fmt.Errorf("failed to get aliases: %w", err)
	}
	detail.Aliases = aliases

	return detail, nil
}
