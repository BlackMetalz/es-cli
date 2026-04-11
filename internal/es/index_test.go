package es

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListIndices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		w.Write([]byte(`[
			{"health":"green","status":"open","index":"test-index","pri":"1","rep":"1","docs.count":"100","pri.store.size":"4.5kb"},
			{"health":"yellow","status":"open","index":"another-index","pri":"3","rep":"0","docs.count":"5000","pri.store.size":"1.2mb"}
		]`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	indices, err := c.ListIndices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(indices) != 2 {
		t.Fatalf("expected 2 indices, got %d", len(indices))
	}
	if indices[0].Name != "test-index" {
		t.Fatalf("expected test-index, got %s", indices[0].Name)
	}
	if indices[0].Health != "green" {
		t.Fatalf("expected green, got %s", indices[0].Health)
	}
	if indices[1].DocsCount != "5000" {
		t.Fatalf("expected 5000, got %s", indices[1].DocsCount)
	}
}

func TestListIndices_Empty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	indices, err := c.ListIndices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(indices) != 0 {
		t.Fatalf("expected 0 indices, got %d", len(indices))
	}
}

func TestCreateIndex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/my-new-index" {
			t.Fatalf("expected /my-new-index, got %s", r.URL.Path)
		}
		w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	err := c.CreateIndex("my-new-index", 3, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCloseIndex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/my-index/_close" {
			t.Fatalf("expected /my-index/_close, got %s", r.URL.Path)
		}
		w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	err := c.CloseIndex("my-index")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteIndex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/my-index" {
			t.Fatalf("expected /my-index, got %s", r.URL.Path)
		}
		w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	err := c.DeleteIndex("my-index")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"52428800", "50.0mb"},
		{"1073741824", "1.0gb"},
		{"1024", "1.0kb"},
		{"500", "500b"},
		{"0", "0b"},
		{"", "-"},
		{"-", "-"},
		{"1099511627776", "1.0tb"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := FormatBytes(tt.input)
			if result != tt.expected {
				t.Errorf("FormatBytes(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseSizeToBytes(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"4.5kb", 4608},
		{"1.2mb", 1258291},
		{"1gb", 1073741824},
		{"500b", 500},
		{"2.5tb", 2748779069440},
		{"0b", 0},
		{"", 0},
		{"-", 0},
		{"1024", 1024},
		{"invalid", 0},
		{"  1.5kb  ", 1536},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseSizeToBytes(tt.input)
			if result != tt.expected {
				t.Errorf("ParseSizeToBytes(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
