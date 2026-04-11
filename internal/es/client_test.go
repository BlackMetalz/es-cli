package es

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:9200/", "elastic", "elastic")
	if c.baseURL != "http://localhost:9200" {
		t.Fatalf("expected trailing slash stripped, got %s", c.baseURL)
	}
	if c.username != "elastic" || c.password != "elastic" {
		t.Fatal("credentials not set")
	}
}

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/test" {
			t.Fatalf("expected /test, got %s", r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "elastic" || pass != "elastic" {
			t.Fatal("basic auth not set correctly")
		}
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	data, err := c.Get("/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `{"ok":true}` {
		t.Fatalf("unexpected response: %s", string(data))
	}
}

func TestClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatal("expected Content-Type: application/json")
		}
		w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	data, err := c.Put("/test-index", `{"settings":{}}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != `{"acknowledged":true}` {
		t.Fatalf("unexpected response: %s", string(data))
	}
}

func TestClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	_, err := c.Delete("/test-index")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"index_not_found"}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	_, err := c.Get("/nonexistent")
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		w.Write([]byte(`{"acknowledged":true}`))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	_, err := c.Post("/test/_close", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
