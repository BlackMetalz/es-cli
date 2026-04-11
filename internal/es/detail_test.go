package es

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetIndexDetail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/test-index/_settings":
			w.Write([]byte(`{"test-index":{"settings":{"index":{"number_of_shards":"1"}}}}`))
		case "/test-index/_mapping":
			w.Write([]byte(`{"test-index":{"mappings":{"properties":{"name":{"type":"text"}}}}}`))
		case "/test-index/_alias":
			w.Write([]byte(`{"test-index":{"aliases":{"my-alias":{}}}}`))
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	detail, err := c.GetIndexDetail("test-index")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.Settings == nil {
		t.Fatal("expected settings")
	}
	if detail.Mappings == nil {
		t.Fatal("expected mappings")
	}
	if detail.Aliases == nil {
		t.Fatal("expected aliases")
	}
}

func TestGetIndexDetail_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"index not found"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	_, err := c.GetIndexDetail("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent index")
	}
}

func TestGetClusterInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Write([]byte(`{"cluster_name":"test-cluster","version":{"number":"8.17.0"}}`))
		case "/_cluster/health":
			w.Write([]byte(`{"status":"green"}`))
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	info, err := c.GetClusterInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Name != "test-cluster" {
		t.Fatalf("expected test-cluster, got %s", info.Name)
	}
	if info.Version != "8.17.0" {
		t.Fatalf("expected 8.17.0, got %s", info.Version)
	}
	if info.Health != "green" {
		t.Fatalf("expected green, got %s", info.Health)
	}
}

func TestOpenIndex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/test-index/_open" {
			w.Write([]byte(`{"acknowledged":true}`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	err := c.OpenIndex("test-index")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
