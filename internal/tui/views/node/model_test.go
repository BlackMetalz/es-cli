package node

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

	if m.Name() != "Nodes" {
		t.Fatalf("expected Nodes, got %s", m.Name())
	}
	if !m.loading {
		t.Fatal("expected loading=true initially")
	}
}

func TestUpdate_NodesLoaded(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)

	nodes := []es.Node{
		{Name: "node-b", IP: "10.0.0.2", CPU: "80", HeapPercent: "50", RAMPercent: "70", Load1m: "1.5", Load5m: "1.2", Load15m: "1.0", NodeRole: "dim", Master: "*", DiskUsedPercent: "40"},
		{Name: "node-a", IP: "10.0.0.1", CPU: "20", HeapPercent: "30", RAMPercent: "40", Load1m: "0.5", Load5m: "0.3", Load15m: "0.2", NodeRole: "dim", Master: "-", DiskUsedPercent: "60"},
	}

	v, _ := m.Update(NodesLoadedMsg{Nodes: nodes})
	updated := v.(*Model)

	if updated.loading {
		t.Fatal("expected loading=false after nodes loaded")
	}
	if updated.err != nil {
		t.Fatalf("unexpected error: %v", updated.err)
	}
	if len(updated.nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(updated.nodes))
	}
	if updated.nodes[0].Name != "node-a" {
		t.Fatalf("expected node-a first (sorted by name ASC), got %s", updated.nodes[0].Name)
	}
}

func TestUpdate_SortByCPU(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.nodes = []es.Node{
		{Name: "low-cpu", CPU: "10"},
		{Name: "high-cpu", CPU: "90"},
		{Name: "mid-cpu", CPU: "50"},
	}

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'C'}}
	v, _ := m.Update(keyMsg)
	updated := v.(*Model)

	if updated.sortField != SortByCPU {
		t.Fatal("expected sort by CPU")
	}
	if updated.filtered[0].Name != "low-cpu" {
		t.Fatalf("expected low-cpu first (ASC), got %s", updated.filtered[0].Name)
	}

	v, _ = updated.Update(keyMsg)
	updated = v.(*Model)
	if updated.filtered[0].Name != "high-cpu" {
		t.Fatalf("expected high-cpu first (DESC), got %s", updated.filtered[0].Name)
	}
}

func TestUpdate_SortByName(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.nodes = []es.Node{
		{Name: "charlie"},
		{Name: "alpha"},
		{Name: "bravo"},
	}

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}}
	v, _ := m.Update(keyMsg)
	updated := v.(*Model)

	if updated.sortField != SortByName {
		t.Fatal("expected sort by name")
	}
	// Default sort is already name ASC, pressing N toggles to DESC
	if updated.filtered[0].Name != "charlie" {
		t.Fatalf("expected charlie first (DESC), got %s", updated.filtered[0].Name)
	}

	v, _ = updated.Update(keyMsg)
	updated = v.(*Model)
	if updated.filtered[0].Name != "alpha" {
		t.Fatalf("expected alpha first (ASC), got %s", updated.filtered[0].Name)
	}
}

func TestUpdate_Maintenance(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.nodes = []es.Node{
		{Name: "node-1"},
	}
	m.filtered = m.nodes

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}}
	v, _ := m.Update(keyMsg)
	updated := v.(*Model)

	pa := updated.PopPendingAction()
	if pa == nil {
		t.Fatal("expected pending action")
	}
	if pa.Type != "set_allocation" {
		t.Fatalf("expected set_allocation, got %s", pa.Type)
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
		t.Fatal("expected command to fetch nodes")
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

	info := m.StatusInfo()
	if !strings.Contains(info, "0 nodes") {
		t.Fatalf("expected '0 nodes', got %s", info)
	}
}
