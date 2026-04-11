package es

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListIndices_RealFormat(t *testing.T) {
	// Simulate real ES response with null values and string sizes
	realResponse := `[
		{"health":"green","status":"close","index":"demo-1","pri":"1","rep":"0","docs.count":null,"pri.store.size":null},
		{"health":"yellow","status":"open","index":"kienlt-index","pri":"1","rep":"1","docs.count":"0","pri.store.size":"249b"},
		{"health":"green","status":"open","index":"demo-2","pri":"1","rep":"0","docs.count":"50000","pri.store.size":"91.8mb"}
	]`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(realResponse))
	}))
	defer server.Close()

	c := NewClient(server.URL, "elastic", "elastic")
	indices, err := c.ListIndices()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(indices) != 3 {
		t.Fatalf("expected 3 indices, got %d", len(indices))
	}

	// demo-1: closed, null values
	if indices[0].DocsCount != "" {
		t.Errorf("demo-1 docs.count: expected empty, got [%s]", indices[0].DocsCount)
	}
	if indices[0].StoreSize != "" {
		t.Errorf("demo-1 store.size: expected empty, got [%s]", indices[0].StoreSize)
	}

	// kienlt-index: small
	if indices[1].StoreSize != "249b" {
		t.Errorf("kienlt-index store.size: expected 249b, got [%s]", indices[1].StoreSize)
	}

	// demo-2: large
	if indices[2].StoreSize != "91.8mb" {
		t.Errorf("demo-2 store.size: expected 91.8mb, got [%s]", indices[2].StoreSize)
	}
	if indices[2].DocsCount != "50000" {
		t.Errorf("demo-2 docs.count: expected 50000, got [%s]", indices[2].DocsCount)
	}

	// Also verify JSON round-trip
	out, _ := json.Marshal(indices[2])
	t.Logf("demo-2 JSON: %s", string(out))
	t.Logf("demo-2 StoreSize: [%s]", indices[2].StoreSize)
}
