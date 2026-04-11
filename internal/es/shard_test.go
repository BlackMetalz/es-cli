package es

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListShards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[
			{"index":"demo-1","shard":"0","prirep":"p","state":"STARTED","docs":"50000","store":"91.8mb","ip":"10.0.0.1","node":"node-1"},
			{"index":"demo-1","shard":"0","prirep":"r","state":"UNASSIGNED","docs":null,"store":null,"ip":null,"node":null}
		]`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	shards, err := c.ListShards()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(shards) != 2 {
		t.Fatalf("expected 2 shards, got %d", len(shards))
	}
	if shards[0].Index != "demo-1" {
		t.Fatalf("expected demo-1, got %s", shards[0].Index)
	}
	if shards[0].PriRep != "p" {
		t.Fatalf("expected p, got %s", shards[0].PriRep)
	}
	if shards[1].State != "UNASSIGNED" {
		t.Fatalf("expected UNASSIGNED, got %s", shards[1].State)
	}
	if shards[1].Node != "" {
		t.Fatalf("expected empty node, got %s", shards[1].Node)
	}
}

func TestListShards_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"failed"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	_, err := c.ListShards()
	if err == nil {
		t.Fatal("expected error")
	}
}
