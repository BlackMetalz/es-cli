package es

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetFieldMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"logs-2024.01": {
				"mappings": {
					"properties": {
						"@timestamp": {"type": "date"},
						"level": {"type": "keyword"},
						"message": {"type": "text"},
						"host": {
							"properties": {
								"name": {"type": "keyword"},
								"ip": {"type": "ip"}
							}
						}
					}
				}
			}
		}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	fields, err := c.GetFieldMapping("logs-*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have: @timestamp, host.ip, host.name, level, message (sorted)
	if len(fields) != 5 {
		t.Fatalf("expected 5 fields, got %d", len(fields))
	}
	if fields[0].Name != "@timestamp" || fields[0].Type != "date" {
		t.Fatalf("expected @timestamp/date, got %s/%s", fields[0].Name, fields[0].Type)
	}
	if fields[1].Name != "host.ip" {
		t.Fatalf("expected host.ip, got %s", fields[1].Name)
	}
}

func TestSearchDocs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"took": 5,
			"hits": {
				"total": {"value": 2, "relation": "eq"},
				"hits": [
					{"_source": {"@timestamp": "2024-01-01T10:00:00Z", "level": "ERROR"}, "sort": [1704103200000, 1]},
					{"_source": {"@timestamp": "2024-01-01T09:00:00Z", "level": "INFO"}, "sort": [1704099600000, 2]}
				]
			}
		}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	result, err := c.SearchDocs("logs-*", "level:ERROR", []string{"@timestamp", "level"}, 100, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 2 {
		t.Fatalf("expected 2 total, got %d", result.Total)
	}
	if result.Took != 5 {
		t.Fatalf("expected took 5, got %d", result.Took)
	}
	if len(result.Hits) != 2 {
		t.Fatalf("expected 2 hits, got %d", len(result.Hits))
	}
	if result.Hits[0].Source["level"] != "ERROR" {
		t.Fatalf("expected ERROR, got %v", result.Hits[0].Source["level"])
	}
	if len(result.Hits[0].Sort) != 2 {
		t.Fatal("expected sort values")
	}
}

func TestSearchDocs_MatchAll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"took":1,"hits":{"total":{"value":0},"hits":[]}}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	result, err := c.SearchDocs("logs-*", "", nil, 50, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 0 {
		t.Fatalf("expected 0, got %d", result.Total)
	}
}

func TestSearchDocs_WithSearchAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"took":1,"hits":{"total":{"value":1},"hits":[
			{"_source":{"level":"WARN"},"sort":[100,3]}
		]}}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	result, err := c.SearchDocs("logs-*", "", nil, 50, []interface{}{float64(200), float64(2)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(result.Hits))
	}
}

func TestGetIndexNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[
			{"health":"green","status":"open","index":"logs-2024.01","pri":"1","rep":"0","docs.count":"100","pri.store.size":"1mb"},
			{"health":"green","status":"open","index":".internal-something","pri":"1","rep":"0","docs.count":"0","pri.store.size":"249b"},
			{"health":"green","status":"open","index":"metrics-2024.01","pri":"1","rep":"0","docs.count":"50","pri.store.size":"500kb"}
		]`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	names, err := c.GetIndexNames()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(names) != 2 {
		t.Fatalf("expected 2 (no system), got %d", len(names))
	}
	if names[0] != "logs-2024.01" {
		t.Fatalf("expected logs-2024.01, got %s", names[0])
	}
}
