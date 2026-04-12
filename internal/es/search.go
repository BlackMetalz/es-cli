package es

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type FieldMapping struct {
	Name string
	Type string // "date", "keyword", "text", "long", "float", etc.
}

type SearchHit struct {
	Source map[string]interface{}
	Sort   []interface{}
}

type SearchResult struct {
	Total int64
	Took  int
	Hits  []SearchHit
}

// GetFieldMapping returns the field mappings for an index pattern.
func (c *Client) GetFieldMapping(index string) ([]FieldMapping, error) {
	data, err := c.Get("/" + index + "/_mapping")
	if err != nil {
		return nil, fmt.Errorf("failed to get mapping for %s: %w", index, err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse mapping: %w", err)
	}

	// Collect fields from all indices matching the pattern
	fieldSet := make(map[string]string) // name → type
	for _, indexData := range raw {
		if idx, ok := indexData.(map[string]interface{}); ok {
			if mappings, ok := idx["mappings"].(map[string]interface{}); ok {
				if props, ok := mappings["properties"].(map[string]interface{}); ok {
					flattenProperties("", props, fieldSet)
				}
			}
		}
	}

	var fields []FieldMapping
	for name, typ := range fieldSet {
		fields = append(fields, FieldMapping{Name: name, Type: typ})
	}
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})

	return fields, nil
}

func flattenProperties(prefix string, props map[string]interface{}, result map[string]string) {
	for name, val := range props {
		fullName := name
		if prefix != "" {
			fullName = prefix + "." + name
		}

		field, ok := val.(map[string]interface{})
		if !ok {
			continue
		}

		if typ, ok := field["type"].(string); ok {
			result[fullName] = typ
		}

		// Recurse into nested properties
		if nested, ok := field["properties"].(map[string]interface{}); ok {
			flattenProperties(fullName, nested, result)
		}
	}
}

// SearchDocs executes a search query against an index.
func (c *Client) SearchDocs(index, query string, fields []string, size int, searchAfter []interface{}) (*SearchResult, error) {
	// Build query
	var queryPart string
	if query == "" {
		queryPart = `"query": {"match_all": {}}`
	} else {
		escaped, _ := json.Marshal(query)
		queryPart = fmt.Sprintf(`"query": {"query_string": {"query": %s, "default_operator": "AND"}}`, string(escaped))
	}

	// Build source filter
	sourcePart := ""
	if len(fields) > 0 {
		fieldJSON, _ := json.Marshal(fields)
		sourcePart = fmt.Sprintf(`, "_source": %s`, string(fieldJSON))
	}

	// Build sort
	sortPart := `"sort": [{"@timestamp": {"order": "desc", "unmapped_type": "date"}}, "_doc"]`

	// Build search_after
	searchAfterPart := ""
	if len(searchAfter) > 0 {
		saJSON, _ := json.Marshal(searchAfter)
		searchAfterPart = fmt.Sprintf(`, "search_after": %s`, string(saJSON))
	}

	body := fmt.Sprintf(`{%s, %s, "size": %d, "track_total_hits": true%s%s}`, queryPart, sortPart, size, sourcePart, searchAfterPart)

	data, err := c.Post("/"+index+"/_search", body)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse search response: %w", err)
	}

	result := &SearchResult{
		Took: jsonInt(raw["took"]),
	}

	if hits, ok := raw["hits"].(map[string]interface{}); ok {
		if total, ok := hits["total"].(map[string]interface{}); ok {
			result.Total = jsonInt64(total["value"])
		}

		if hitList, ok := hits["hits"].([]interface{}); ok {
			for _, h := range hitList {
				hit, ok := h.(map[string]interface{})
				if !ok {
					continue
				}
				searchHit := SearchHit{}
				if src, ok := hit["_source"].(map[string]interface{}); ok {
					searchHit.Source = src
				}
				if sortVals, ok := hit["sort"].([]interface{}); ok {
					searchHit.Sort = sortVals
				}
				result.Hits = append(result.Hits, searchHit)
			}
		}
	}

	return result, nil
}

// GetIndexNames returns deduplicated index names from ListIndices (for pattern selection).
func (c *Client) GetIndexNames() ([]string, error) {
	indices, err := c.ListIndices()
	if err != nil {
		return nil, err
	}
	var names []string
	for _, idx := range indices {
		if !strings.HasPrefix(idx.Name, ".") {
			names = append(names, idx.Name)
		}
	}
	sort.Strings(names)
	return names, nil
}
