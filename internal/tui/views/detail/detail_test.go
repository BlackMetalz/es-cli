package detail

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kienlt/es-cli/internal/es"
)

func newTestClient() *es.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/_settings"):
			w.Write([]byte(`{"idx":{"settings":{"number_of_shards":"1"}}}`))
		case strings.HasSuffix(r.URL.Path, "/_mapping"):
			w.Write([]byte(`{"idx":{"mappings":{}}}`))
		case strings.HasSuffix(r.URL.Path, "/_alias"):
			w.Write([]byte(`{"idx":{"aliases":{}}}`))
		}
	}))
	return es.NewClient(server.URL, "elastic", "elastic")
}

func TestNew(t *testing.T) {
	client := newTestClient()
	m := New(client, "test-index")

	if m.Name() != "Index: test-index" {
		t.Fatalf("expected 'Index: test-index', got %s", m.Name())
	}
	if !m.loading {
		t.Fatal("expected loading=true")
	}
}

func TestUpdate_DetailLoaded(t *testing.T) {
	client := newTestClient()
	m := New(client, "test-index")
	m.SetSize(80, 24)

	detail := &es.IndexDetail{
		Settings: json.RawMessage(`{"settings":{"number_of_shards":"1"}}`),
		Mappings: json.RawMessage(`{"mappings":{}}`),
		Aliases:  json.RawMessage(`{"aliases":{}}`),
	}

	v, _ := m.Update(DetailLoadedMsg{Detail: detail})
	updated := v.(*Model)

	if updated.loading {
		t.Fatal("expected loading=false")
	}
	if updated.detail == nil {
		t.Fatal("expected detail set")
	}
}

func TestUpdate_Error(t *testing.T) {
	client := newTestClient()
	m := New(client, "test-index")

	v, _ := m.Update(ErrorMsg{Err: nil})
	updated := v.(*Model)

	if updated.loading {
		t.Fatal("expected loading=false")
	}
}

func TestUpdate_TabSwitch(t *testing.T) {
	client := newTestClient()
	m := New(client, "test-index")
	m.loading = false
	m.detail = &es.IndexDetail{
		Settings: json.RawMessage(`{}`),
		Mappings: json.RawMessage(`{}`),
		Aliases:  json.RawMessage(`{}`),
	}
	m.SetSize(80, 24)

	if m.tab != TabSettings {
		t.Fatal("expected settings tab initially")
	}

	// Tab → mappings
	v, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated := v.(*Model)
	if updated.tab != TabMappings {
		t.Fatalf("expected mappings tab, got %d", updated.tab)
	}

	// Tab → aliases
	v, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated = v.(*Model)
	if updated.tab != TabAliases {
		t.Fatalf("expected aliases tab, got %d", updated.tab)
	}

	// Tab → wraps to settings
	v, _ = updated.Update(tea.KeyMsg{Type: tea.KeyTab})
	updated = v.(*Model)
	if updated.tab != TabSettings {
		t.Fatalf("expected settings tab, got %d", updated.tab)
	}

	// Shift+Tab → aliases (backwards)
	v, _ = updated.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	updated = v.(*Model)
	if updated.tab != TabAliases {
		t.Fatalf("expected aliases tab, got %d", updated.tab)
	}
}

func TestUpdate_EscGoesBack(t *testing.T) {
	client := newTestClient()
	m := New(client, "test-index")
	m.loading = false

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command from esc")
	}

	msg := cmd()
	if _, ok := msg.(GoBackMsg); !ok {
		t.Fatal("expected GoBackMsg")
	}
}

func TestView_Loading(t *testing.T) {
	client := newTestClient()
	m := New(client, "test-index")

	view := m.View()
	if !strings.Contains(view, "Loading") {
		t.Fatal("expected loading text")
	}
}

func TestView_WithData(t *testing.T) {
	client := newTestClient()
	m := New(client, "test-index")
	m.loading = false
	m.detail = &es.IndexDetail{
		Settings: json.RawMessage(`{"number_of_shards":"1"}`),
		Mappings: json.RawMessage(`{}`),
		Aliases:  json.RawMessage(`{}`),
	}
	m.SetSize(80, 24)
	m.updateViewport()

	view := m.View()
	if !strings.Contains(view, "Settings") {
		t.Fatal("expected Settings tab in view")
	}
}

func TestHelpGroups(t *testing.T) {
	client := newTestClient()
	m := New(client, "test-index")
	groups := m.HelpGroups()

	if len(groups) != 1 {
		t.Fatalf("expected 1 help group, got %d", len(groups))
	}
	if groups[0].Title != "Navigation" {
		t.Fatalf("expected Navigation, got %s", groups[0].Title)
	}
}

func TestIsInputMode(t *testing.T) {
	client := newTestClient()
	m := New(client, "test-index")
	if m.IsInputMode() {
		t.Fatal("detail view should not be in input mode")
	}
}

func TestPopPendingAction(t *testing.T) {
	client := newTestClient()
	m := New(client, "test-index")
	if m.PopPendingAction() != nil {
		t.Fatal("expected nil pending action")
	}
}

func TestStatusInfo(t *testing.T) {
	client := newTestClient()
	m := New(client, "test-index")
	if m.StatusInfo() != "test-index" {
		t.Fatalf("expected test-index, got %s", m.StatusInfo())
	}
}
