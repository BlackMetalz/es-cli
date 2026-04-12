package query

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kienlt/es-cli/internal/es"
)

func newTestClient(handler http.Handler) *es.Client {
	server := httptest.NewServer(handler)
	return es.NewClient(server.URL, "elastic", "elastic")
}

func TestNew(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)

	if m.Name() != "Query" {
		t.Fatalf("expected Query, got %s", m.Name())
	}
	if !m.selectingIndex {
		t.Fatal("expected selectingIndex=true initially")
	}
	if !m.loading {
		t.Fatal("expected loading=true initially")
	}
}

func TestUpdate_IndicesLoaded(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)

	names := []string{"logs-2024", "metrics-2024", "events"}
	v, _ := m.Update(IndicesLoadedMsg{Names: names})
	updated := v.(*Model)

	if updated.loading {
		t.Fatal("expected loading=false after indices loaded")
	}
	if len(updated.indices) != 3 {
		t.Fatalf("expected 3 indices, got %d", len(updated.indices))
	}
	if len(updated.filteredIdx) != 3 {
		t.Fatalf("expected 3 filtered indices, got %d", len(updated.filteredIdx))
	}
}

func TestUpdate_SelectIndex(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.selectingIndex = true
	m.indices = []string{"logs-2024", "metrics-2024"}
	m.filteredIdx = m.indices
	m.indexCursor = 0

	// Press enter to select index
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	v, cmd := m.Update(keyMsg)
	updated := v.(*Model)

	if updated.selectingIndex {
		t.Fatal("expected selectingIndex=false after selection")
	}
	if updated.index != "logs-2024" {
		t.Fatalf("expected logs-2024, got %s", updated.index)
	}
	if !updated.loading {
		t.Fatal("expected loading=true while fetching mapping")
	}
	if cmd == nil {
		t.Fatal("expected command to fetch mapping")
	}
}

func TestUpdate_MappingLoaded(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.selectingIndex = false
	m.index = "logs-2024"
	m.loading = true
	m.SetSize(120, 40)

	fields := []es.FieldMapping{
		{Name: "@timestamp", Type: "date"},
		{Name: "message", Type: "text"},
		{Name: "level", Type: "keyword"},
		{Name: "host.name", Type: "keyword"},
	}

	v, cmd := m.Update(MappingLoadedMsg{Fields: fields})
	updated := v.(*Model)

	if updated.loading {
		t.Fatal("expected loading=false after mapping loaded")
	}
	if len(updated.allFields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(updated.allFields))
	}
	if len(updated.columns) == 0 {
		t.Fatal("expected columns to be auto-detected")
	}
	// @timestamp should be first column
	if updated.columns[0] != "@timestamp" {
		t.Fatalf("expected @timestamp as first column, got %s", updated.columns[0])
	}
	if cmd == nil {
		t.Fatal("expected command to execute search")
	}
}

func TestUpdate_SearchResult(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.selectingIndex = false
	m.index = "logs-2024"
	m.columns = []string{"@timestamp", "message"}
	m.SetSize(120, 40)

	result := &es.SearchResult{
		Total: 42,
		Took:  5,
		Hits: []es.SearchHit{
			{Source: map[string]interface{}{"@timestamp": "2024-01-01", "message": "hello"}, Sort: []interface{}{1704067200000, 1}},
			{Source: map[string]interface{}{"@timestamp": "2024-01-02", "message": "world"}, Sort: []interface{}{1704153600000, 2}},
		},
	}

	v, _ := m.Update(SearchResultMsg{Result: result})
	updated := v.(*Model)

	if updated.total != 42 {
		t.Fatalf("expected total=42, got %d", updated.total)
	}
	if len(updated.hits) != 2 {
		t.Fatalf("expected 2 hits, got %d", len(updated.hits))
	}
}

func TestUpdate_FollowToggle(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.selectingIndex = false
	m.index = "logs-2024"
	m.loading = false

	// Press f to cycle follow
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'f'}}

	v, cmd := m.Update(keyMsg)
	updated := v.(*Model)
	if updated.followInterval != 1*time.Second {
		t.Fatalf("expected 1s follow, got %s", updated.followInterval)
	}
	if cmd == nil {
		t.Fatal("expected tick command when follow enabled")
	}

	// Press f again
	v, cmd = updated.Update(keyMsg)
	updated = v.(*Model)
	if updated.followInterval != 2*time.Second {
		t.Fatalf("expected 2s follow, got %s", updated.followInterval)
	}

	// Cycle through all intervals back to 0
	for updated.followInterval != 0 {
		v, _ = updated.Update(keyMsg)
		updated = v.(*Model)
	}
	if updated.followIdx != 0 {
		t.Fatalf("expected followIdx=0 after cycling back, got %d", updated.followIdx)
	}
}

func TestUpdate_ErrorMsg(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)

	v, _ := m.Update(ErrorMsg{Err: fmt.Errorf("connection refused")})
	updated := v.(*Model)

	if updated.err == nil {
		t.Fatal("expected error to be set")
	}
	if updated.loading {
		t.Fatal("expected loading=false")
	}
}

func TestView_SelectingIndex(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.selectingIndex = true
	m.indices = []string{"logs-2024", "metrics-2024"}
	m.filteredIdx = m.indices
	m.SetSize(80, 30)

	view := m.View()
	if !strings.Contains(view, "Select Index") {
		t.Fatal("expected 'Select Index' in view")
	}
	if !strings.Contains(view, "logs-2024") {
		t.Fatal("expected index names in view")
	}
}

func TestHelpGroups(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	groups := m.HelpGroups()

	if len(groups) != 2 {
		t.Fatalf("expected 2 help groups, got %d", len(groups))
	}
	if groups[0].Title != "Query" {
		t.Fatalf("expected Query group first, got %s", groups[0].Title)
	}
	if groups[1].Title != "General" {
		t.Fatalf("expected General group second, got %s", groups[1].Title)
	}
}

func TestIsInputMode(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)

	if m.IsInputMode() {
		t.Fatal("expected not in input mode initially")
	}

	m.editing = true
	if !m.IsInputMode() {
		t.Fatal("expected input mode when editing")
	}

	m.editing = false
	m.indexSearching = true
	if !m.IsInputMode() {
		t.Fatal("expected input mode when index searching")
	}
}

func TestStatusInfo(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)

	info := m.StatusInfo()
	if info != "selecting index" {
		t.Fatalf("expected 'selecting index', got %s", info)
	}

	m.index = "logs-2024"
	info = m.StatusInfo()
	if !strings.Contains(info, "logs-2024") {
		t.Fatalf("expected 'logs-2024' in status, got %s", info)
	}

	m.followInterval = 5 * time.Second
	info = m.StatusInfo()
	if !strings.Contains(info, "live") {
		t.Fatalf("expected 'live' in status, got %s", info)
	}
}

func TestGetField(t *testing.T) {
	source := map[string]interface{}{
		"message": "hello",
		"host": map[string]interface{}{
			"name": "server1",
		},
	}

	if v := getField(source, "message"); v != "hello" {
		t.Fatalf("expected 'hello', got '%s'", v)
	}
	if v := getField(source, "host.name"); v != "server1" {
		t.Fatalf("expected 'server1', got '%s'", v)
	}
	if v := getField(source, "missing"); v != "" {
		t.Fatalf("expected empty, got '%s'", v)
	}
	if v := getField(source, "host.missing"); v != "" {
		t.Fatalf("expected empty, got '%s'", v)
	}
}

func TestDefaultColumns(t *testing.T) {
	fields := []es.FieldMapping{
		{Name: "@timestamp", Type: "date"},
		{Name: "message", Type: "text"},
		{Name: "level", Type: "keyword"},
		{Name: "count", Type: "long"},
	}
	cols := defaultColumns(fields)
	if len(cols) == 0 {
		t.Fatal("expected at least one column")
	}
	if cols[0] != "@timestamp" {
		t.Fatalf("expected @timestamp first, got %s", cols[0])
	}
	// Should include message and level (text/keyword) but not count (long)
	hasMessage := false
	hasLevel := false
	for _, c := range cols {
		if c == "message" {
			hasMessage = true
		}
		if c == "level" {
			hasLevel = true
		}
	}
	if !hasMessage {
		t.Fatal("expected message in default columns")
	}
	if !hasLevel {
		t.Fatal("expected level in default columns")
	}
}
