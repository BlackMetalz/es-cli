package es

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListNodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[
			{"name":"node-1","ip":"10.0.0.1","heap.percent":"45","ram.percent":"67","cpu":"12","load_1m":"0.5","load_5m":"0.3","load_15m":"0.2","node.role":"cdfhilmrstw","master":"*","disk.used_percent":"55.3"},
			{"name":"node-2","ip":"10.0.0.2","heap.percent":"60","ram.percent":"80","cpu":"25","load_1m":"1.2","load_5m":"0.8","load_15m":"0.5","node.role":"cdfhilmrstw","master":"-","disk.used_percent":"70.1"}
		]`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	nodes, err := c.ListNodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Name != "node-1" {
		t.Fatalf("expected node-1, got %s", nodes[0].Name)
	}
	if nodes[0].Master != "*" {
		t.Fatalf("expected master *, got %s", nodes[0].Master)
	}
	if nodes[1].HeapPercent != "60" {
		t.Fatalf("expected 60, got %s", nodes[1].HeapPercent)
	}
}

func TestListNodes_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"cluster error"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	_, err := c.ListNodes()
	if err == nil {
		t.Fatal("expected error")
	}
}
