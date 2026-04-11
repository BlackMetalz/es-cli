package es

import (
	"encoding/json"
	"fmt"
	"strings"
)

type IndexTemplate struct {
	Name          string
	IndexPatterns string
	Priority      int
	Shards        string
	Replicas      string
	ILMPolicy     string
	Managed       bool
}

func (c *Client) ListIndexTemplates() ([]IndexTemplate, error) {
	data, err := c.Get("/_index_template")
	if err != nil {
		return nil, fmt.Errorf("failed to list index templates: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse index templates response: %w", err)
	}

	templatesList, ok := raw["index_templates"].([]interface{})
	if !ok {
		return nil, nil
	}

	var templates []IndexTemplate
	for _, item := range templatesList {
		entry, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		t := IndexTemplate{
			Name: JsonStr(entry["name"]),
		}

		if it, ok := entry["index_template"].(map[string]interface{}); ok {
			// Index patterns
			if patterns, ok := it["index_patterns"].([]interface{}); ok {
				var pats []string
				for _, p := range patterns {
					pats = append(pats, JsonStr(p))
				}
				t.IndexPatterns = strings.Join(pats, ", ")
			}

			t.Priority = jsonInt(it["priority"])

			// Check managed flag
			if meta, ok := it["_meta"].(map[string]interface{}); ok {
				if managed, ok := meta["managed"].(bool); ok && managed {
					t.Managed = true
				}
			}

			// Settings from template
			if tmpl, ok := it["template"].(map[string]interface{}); ok {
				if settings, ok := tmpl["settings"].(map[string]interface{}); ok {
					if idx, ok := settings["index"].(map[string]interface{}); ok {
						t.Shards = JsonStr(idx["number_of_shards"])
						t.Replicas = JsonStr(idx["number_of_replicas"])
						if lifecycle, ok := idx["lifecycle"].(map[string]interface{}); ok {
							t.ILMPolicy = JsonStr(lifecycle["name"])
						}
					}
					// Some templates have flat settings
					if t.Shards == "" {
						t.Shards = JsonStr(settings["number_of_shards"])
					}
					if t.Replicas == "" {
						t.Replicas = JsonStr(settings["number_of_replicas"])
					}
				}
			}
		}

		templates = append(templates, t)
	}

	return templates, nil
}

func (c *Client) CreateIndexTemplate(name string, body string) error {
	_, err := c.Put("/_index_template/"+name, body)
	if err != nil {
		return fmt.Errorf("failed to create index template %s: %w", name, err)
	}
	return nil
}

func (c *Client) GetIndexTemplate(name string) ([]byte, error) {
	data, err := c.Get("/_index_template/" + name)
	if err != nil {
		return nil, fmt.Errorf("failed to get index template %s: %w", name, err)
	}
	return data, nil
}

func (c *Client) DeleteIndexTemplate(name string) error {
	_, err := c.Delete("/_index_template/" + name)
	if err != nil {
		return fmt.Errorf("failed to delete index template %s: %w", name, err)
	}
	return nil
}
