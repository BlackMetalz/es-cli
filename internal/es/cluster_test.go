package es

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAllocationSetting_Default(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"persistent":{},"transient":{}}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	val, err := c.GetAllocationSetting()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty (default), got %s", val)
	}
}

func TestGetAllocationSetting_Primaries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"persistent":{},"transient":{"cluster.routing.allocation.enable":"primaries"}}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	val, err := c.GetAllocationSetting()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "primaries" {
		t.Fatalf("expected primaries, got %s", val)
	}
}

func TestGetAllocationSetting_Persistent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"persistent":{"cluster.routing.allocation.enable":"none"},"transient":{}}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	val, err := c.GetAllocationSetting()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "none" {
		t.Fatalf("expected none, got %s", val)
	}
}

func TestSetAllocationSetting(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])
		w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	err := c.SetAllocationSetting("primaries")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedBody != `{"transient":{"cluster.routing.allocation.enable":"primaries"}}` {
		t.Fatalf("unexpected body: %s", receivedBody)
	}
}

func TestSetAllocationSetting_Reset(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])
		w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	err := c.SetAllocationSetting("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedBody != `{"transient":{"cluster.routing.allocation.enable":null}}` {
		t.Fatalf("unexpected body: %s", receivedBody)
	}
}
