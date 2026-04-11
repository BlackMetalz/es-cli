package index

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

	if m.Name() != "Indices" {
		t.Fatalf("expected Indices, got %s", m.Name())
	}
	if !m.loading {
		t.Fatal("expected loading=true initially")
	}
}

func TestUpdate_IndicesLoaded(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)

	indices := []es.Index{
		{Health: "green", Status: "open", Name: "b-index", Pri: "1", Rep: "1", DocsCount: "100", StoreSize: "4.5kb"},
		{Health: "yellow", Status: "open", Name: "a-index", Pri: "3", Rep: "0", DocsCount: "5000", StoreSize: "1.2mb"},
	}

	v, _ := m.Update(IndicesLoadedMsg{Indices: indices})
	updated := v.(*Model)

	if updated.loading {
		t.Fatal("expected loading=false after indices loaded")
	}
	if updated.err != nil {
		t.Fatalf("unexpected error: %v", updated.err)
	}
	if len(updated.indices) != 2 {
		t.Fatalf("expected 2 indices, got %d", len(updated.indices))
	}
	if updated.indices[0].Name != "a-index" {
		t.Fatalf("expected a-index first (sorted by name), got %s", updated.indices[0].Name)
	}
}

func TestUpdate_SortBySize(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.indices = []es.Index{
		{Name: "small", StoreSize: "1024"},
		{Name: "large", StoreSize: "1073741824"},
		{Name: "medium", StoreSize: "1048576"},
	}

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}}
	v, _ := m.Update(keyMsg)
	updated := v.(*Model)

	if updated.sortField != SortBySize {
		t.Fatal("expected sort by size")
	}
	if updated.indices[0].Name != "small" {
		t.Fatalf("expected small first (ASC), got %s", updated.indices[0].Name)
	}

	v, _ = updated.Update(keyMsg)
	updated = v.(*Model)
	if updated.indices[0].Name != "large" {
		t.Fatalf("expected large first (DESC), got %s", updated.indices[0].Name)
	}
}

func TestUpdate_SortByDocs(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.indices = []es.Index{
		{Name: "few", DocsCount: "10"},
		{Name: "many", DocsCount: "50000"},
		{Name: "some", DocsCount: "1000"},
	}

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}}
	v, _ := m.Update(keyMsg)
	updated := v.(*Model)

	if updated.sortField != SortByDocs {
		t.Fatal("expected sort by docs")
	}
	if updated.indices[0].Name != "few" {
		t.Fatalf("expected few first (ASC), got %s", updated.indices[0].Name)
	}

	v, _ = updated.Update(keyMsg)
	updated = v.(*Model)
	if updated.indices[0].Name != "many" {
		t.Fatalf("expected many first (DESC), got %s", updated.indices[0].Name)
	}
}

func TestUpdate_SortByName(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.indices = []es.Index{
		{Name: "charlie"},
		{Name: "alpha"},
		{Name: "bravo"},
	}

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'I'}}
	v, _ := m.Update(keyMsg)
	updated := v.(*Model)

	if updated.sortField != SortByName {
		t.Fatal("expected sort by name")
	}
	if updated.indices[0].Name != "charlie" {
		t.Fatalf("expected charlie first (DESC), got %s", updated.indices[0].Name)
	}

	v, _ = updated.Update(keyMsg)
	updated = v.(*Model)
	if updated.indices[0].Name != "alpha" {
		t.Fatalf("expected alpha first (ASC), got %s", updated.indices[0].Name)
	}
}

func TestUpdate_DeleteConfirm(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.indices = []es.Index{
		{Name: "test-index"},
	}
	m.filtered = m.indices
	m.updateTable()
	m.table.SetCursor(0)

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	v, _ := m.Update(keyMsg)
	updated := v.(*Model)

	pa := updated.PopPendingAction()
	if pa == nil {
		t.Fatal("expected pending action")
	}
	if pa.Type != "delete" {
		t.Fatalf("expected delete, got %s", pa.Type)
	}
	if pa.Index != "test-index" {
		t.Fatalf("expected test-index, got %s", pa.Index)
	}
}

func TestUpdate_ToggleOpenClose(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.showAll = true
	m.indices = []es.Index{
		{Name: "my-index", Status: "open"},
	}
	m.filtered = m.indices
	m.updateTable()
	m.table.SetCursor(0)

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}
	v, _ := m.Update(keyMsg)
	updated := v.(*Model)

	pa := updated.PopPendingAction()
	if pa == nil {
		t.Fatal("expected pending action")
	}
	if pa.Type != "close" {
		t.Fatalf("expected close, got %s", pa.Type)
	}
}

func TestPopPendingAction_ClearsAction(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)

	// No pending action
	if pa := m.PopPendingAction(); pa != nil {
		t.Fatal("expected nil pending action")
	}
}

func TestUpdate_Refresh(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
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
		t.Fatal("expected command to fetch indices")
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

func TestView_Error(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.err = fmt.Errorf("connection refused")

	view := m.View()
	if !strings.Contains(view, "Error") {
		t.Fatal("expected error text")
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

	info := m.StatusInfo()
	if !strings.Contains(info, "hiding system indices") {
		t.Fatalf("expected 'hiding system indices', got %s", info)
	}
}

func TestNaturalSort(t *testing.T) {
	if naturalCompare("demo-2", "demo-10") >= 0 {
		t.Fatal("expected demo-2 < demo-10")
	}
	if naturalCompare("demo-10", "demo-2") <= 0 {
		t.Fatal("expected demo-10 > demo-2")
	}
	if naturalCompare("abc", "abc") != 0 {
		t.Fatal("expected equal")
	}
}
