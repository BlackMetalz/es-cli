package template

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

	if m.Name() != "Index Templates" {
		t.Fatalf("expected Index Templates, got %s", m.Name())
	}
	if !m.loading {
		t.Fatal("expected loading=true initially")
	}
}

func TestUpdate_TemplatesLoaded(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)

	templates := []es.IndexTemplate{
		{Name: "b-template", IndexPatterns: "b-*", Priority: 10, Shards: "1", Replicas: "1"},
		{Name: "a-template", IndexPatterns: "a-*", Priority: 5, Shards: "3", Replicas: "0"},
	}

	v, _ := m.Update(TemplatesLoadedMsg{Templates: templates})
	updated := v.(*Model)

	if updated.loading {
		t.Fatal("expected loading=false after templates loaded")
	}
	if updated.err != nil {
		t.Fatalf("unexpected error: %v", updated.err)
	}
	if len(updated.templates) != 2 {
		t.Fatalf("expected 2 templates, got %d", len(updated.templates))
	}
	if updated.templates[0].Name != "a-template" {
		t.Fatalf("expected a-template first (sorted by name), got %s", updated.templates[0].Name)
	}
}

func TestUpdate_SortByName(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.templates = []es.IndexTemplate{
		{Name: "charlie"},
		{Name: "alpha"},
		{Name: "bravo"},
	}

	// First press: already sorting by name ASC, so toggles to DESC
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}
	v, _ := m.Update(keyMsg)
	updated := v.(*Model)

	if updated.sortField != SortByName {
		t.Fatal("expected sort by name")
	}
	if updated.templates[0].Name != "charlie" {
		t.Fatalf("expected charlie first (DESC), got %s", updated.templates[0].Name)
	}

	// Second press: toggles back to ASC
	v, _ = updated.Update(keyMsg)
	updated = v.(*Model)
	if updated.templates[0].Name != "alpha" {
		t.Fatalf("expected alpha first (ASC), got %s", updated.templates[0].Name)
	}
}

func TestUpdate_Delete(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.templates = []es.IndexTemplate{
		{Name: "test-template"},
	}
	m.filtered = m.templates
	m.updateTable()
	m.table.SetCursor(0)

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	v, _ := m.Update(keyMsg)
	updated := v.(*Model)

	pa := updated.PopPendingAction()
	if pa == nil {
		t.Fatal("expected pending action")
	}
	if pa.Type != "delete_template" {
		t.Fatalf("expected delete_template, got %s", pa.Type)
	}
	if pa.Index != "test-template" {
		t.Fatalf("expected test-template, got %s", pa.Index)
	}
}

func TestUpdate_Refresh(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"index_templates":[]}`))
	}))
	m := New(client)
	m.loading = false

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	v, cmd := m.Update(keyMsg)
	updated := v.(*Model)

	if !updated.loading {
		t.Fatal("expected loading=true after refresh")
	}
	if cmd == nil {
		t.Fatal("expected command to fetch templates")
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

func TestView_Loading(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	view := m.View()

	if !strings.Contains(view, "Loading") {
		t.Fatal("expected loading text")
	}
}

func TestHelpGroups(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	groups := m.HelpGroups()

	if len(groups) != 3 {
		t.Fatalf("expected 3 help groups, got %d", len(groups))
	}
	if groups[0].Title != "Sort" {
		t.Fatalf("expected Sort group first, got %s", groups[0].Title)
	}
}

func TestSetSize(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.SetSize(120, 40)

	if m.width != 120 {
		t.Fatalf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Fatalf("expected height 40, got %d", m.height)
	}
}

func TestIsInputMode(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)

	if m.IsInputMode() {
		t.Fatal("expected not in input mode initially")
	}
}

func TestStatusInfo(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.filtered = []es.IndexTemplate{}

	info := m.StatusInfo()
	if !strings.Contains(info, "templates") {
		t.Fatalf("expected 'templates' in status info, got %s", info)
	}
}
