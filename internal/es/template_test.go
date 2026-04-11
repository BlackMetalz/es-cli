package es

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListIndexTemplates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{
			"index_templates": [
				{
					"name": "logs-template",
					"index_template": {
						"index_patterns": ["logs-*", "app-logs-*"],
						"priority": 100,
						"template": {
							"settings": {
								"index": {
									"number_of_shards": "2",
									"number_of_replicas": "1",
									"lifecycle": {"name": "logs-policy"}
								}
							}
						}
					}
				},
				{
					"name": "metrics-template",
					"index_template": {
						"index_patterns": ["metrics-*"],
						"priority": 50,
						"template": {
							"settings": {
								"index": {
									"number_of_shards": "1",
									"number_of_replicas": "0"
								}
							}
						}
					}
				}
			]
		}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	templates, err := c.ListIndexTemplates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(templates))
	}
	if templates[0].Name != "logs-template" {
		t.Fatalf("expected logs-template, got %s", templates[0].Name)
	}
	if templates[0].IndexPatterns != "logs-*, app-logs-*" {
		t.Fatalf("expected 'logs-*, app-logs-*', got '%s'", templates[0].IndexPatterns)
	}
	if templates[0].Priority != 100 {
		t.Fatalf("expected priority 100, got %d", templates[0].Priority)
	}
	if templates[0].Shards != "2" {
		t.Fatalf("expected 2 shards, got %s", templates[0].Shards)
	}
	if templates[0].ILMPolicy != "logs-policy" {
		t.Fatalf("expected logs-policy, got %s", templates[0].ILMPolicy)
	}
	if templates[1].ILMPolicy != "" {
		t.Fatalf("expected empty ILM policy, got %s", templates[1].ILMPolicy)
	}
}

func TestListIndexTemplates_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"index_templates":[]}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	templates, err := c.ListIndexTemplates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(templates) != 0 {
		t.Fatalf("expected 0, got %d", len(templates))
	}
}

func TestCreateIndexTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	err := c.CreateIndexTemplate("test-template", `{"index_patterns":["test-*"]}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteIndexTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	err := c.DeleteIndexTemplate("test-template")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
