package es

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetDashboardData(t *testing.T) {
	uptimeMs := int64((3 * time.Hour).Milliseconds())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Write([]byte(`{"cluster_name":"test-cluster","version":{"number":"8.17.0"}}`))
		case "/_cluster/stats":
			w.Write([]byte(fmt.Sprintf(`{
				"status":"green",
				"indices":{
					"count":45,
					"docs":{"count":123456},
					"store":{"size_in_bytes":5368709120},
					"shards":{"total":90,"primaries":45}
				},
				"nodes":{
					"count":{"total":3},
					"jvm":{"mem":{"heap_used_in_bytes":536870912,"heap_max_in_bytes":1073741824}},
					"fs":{"total_in_bytes":107374182400,"available_in_bytes":53687091200}
				}
			}`)))
		case "/_nodes/stats/jvm":
			w.Write([]byte(fmt.Sprintf(`{
				"nodes":{
					"node1":{"jvm":{"uptime_in_millis":%d}},
					"node2":{"jvm":{"uptime_in_millis":%d}}
				}
			}`, uptimeMs, uptimeMs-60000)))
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	d, err := c.GetDashboardData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if d.Version != "8.17.0" {
		t.Fatalf("expected 8.17.0, got %s", d.Version)
	}
	if d.Health != "green" {
		t.Fatalf("expected green, got %s", d.Health)
	}
	if d.NodeCount != 3 {
		t.Fatalf("expected 3 nodes, got %d", d.NodeCount)
	}
	if d.IndexCount != 45 {
		t.Fatalf("expected 45 indices, got %d", d.IndexCount)
	}
	if d.DocCount != 123456 {
		t.Fatalf("expected 123456 docs, got %d", d.DocCount)
	}
	if d.PrimaryShards != 45 {
		t.Fatalf("expected 45 primary shards, got %d", d.PrimaryShards)
	}
	if d.ReplicaShards != 45 {
		t.Fatalf("expected 45 replica shards, got %d", d.ReplicaShards)
	}
	// Uptime should be ~3 hours
	if d.Uptime < 2*time.Hour || d.Uptime > 4*time.Hour {
		t.Fatalf("expected ~3h uptime, got %v", d.Uptime)
	}
	if d.HeapUsedBytes == 0 || d.HeapMaxBytes == 0 {
		t.Fatal("expected heap values")
	}
	if d.DiskAvailBytes == 0 || d.DiskTotalBytes == 0 {
		t.Fatal("expected disk values")
	}
	if d.HeapUsedPercent() < 40 || d.HeapUsedPercent() > 60 {
		t.Fatalf("expected ~50%% heap, got %.2f%%", d.HeapUsedPercent())
	}
}

func TestGetDashboardData_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"cluster error"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	_, err := c.GetDashboardData()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetDashboardData_ZeroIndices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Write([]byte(`{"version":{"number":"8.17.0"}}`))
		case "/_cluster/stats":
			w.Write([]byte(`{"status":"green","indices":{"count":0,"docs":{"count":0},"store":{"size_in_bytes":0},"shards":{"total":0,"primaries":0}},"nodes":{"count":{"total":1},"jvm":{"mem":{"heap_used_in_bytes":0,"heap_max_in_bytes":0}},"fs":{"total_in_bytes":0,"available_in_bytes":0}}}`))
		case "/_nodes/stats/jvm":
			w.Write([]byte(`{"nodes":{}}`))
		}
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	d, err := c.GetDashboardData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.IndexCount != 0 {
		t.Fatalf("expected 0 indices, got %d", d.IndexCount)
	}
	if d.Uptime != 0 {
		t.Fatalf("expected 0 uptime, got %v", d.Uptime)
	}
}
