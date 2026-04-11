package ilm

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

	if m.Name() != "ILM Policies" {
		t.Fatalf("expected ILM Policies, got %s", m.Name())
	}
	if !m.loading {
		t.Fatal("expected loading=true initially")
	}
}

func TestUpdate_PoliciesLoaded(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)

	policies := []es.ILMPolicy{
		{Name: "beta-policy", Version: 1, Phases: map[string]es.ILMPhase{
			"hot": {MinAge: "0ms"},
		}},
		{Name: "alpha-policy", Version: 2, Phases: map[string]es.ILMPhase{
			"hot":  {MinAge: "0ms"},
			"warm": {MinAge: "30d"},
		}},
	}

	v, _ := m.Update(PoliciesLoadedMsg{Policies: policies})
	updated := v.(*Model)

	if updated.loading {
		t.Fatal("expected loading=false after policies loaded")
	}
	if updated.err != nil {
		t.Fatalf("unexpected error: %v", updated.err)
	}
	if len(updated.policies) != 2 {
		t.Fatalf("expected 2 policies, got %d", len(updated.policies))
	}
	if updated.policies[0].Name != "alpha-policy" {
		t.Fatalf("expected alpha-policy first (sorted by name), got %s", updated.policies[0].Name)
	}
}

func TestUpdate_SortByName(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.policies = []es.ILMPolicy{
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
	if updated.policies[0].Name != "charlie" {
		t.Fatalf("expected charlie first (DESC), got %s", updated.policies[0].Name)
	}

	// Second press: toggles back to ASC
	v, _ = updated.Update(keyMsg)
	updated = v.(*Model)
	if updated.policies[0].Name != "alpha" {
		t.Fatalf("expected alpha first (ASC), got %s", updated.policies[0].Name)
	}
}

func TestUpdate_Delete(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.policies = []es.ILMPolicy{
		{Name: "test-policy"},
	}
	m.filtered = m.policies
	m.updateTable()
	m.table.SetCursor(0)

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	v, _ := m.Update(keyMsg)
	updated := v.(*Model)

	pa := updated.PopPendingAction()
	if pa == nil {
		t.Fatal("expected pending action")
	}
	if pa.Type != "delete_ilm" {
		t.Fatalf("expected delete_ilm, got %s", pa.Type)
	}
	if pa.Index != "test-policy" {
		t.Fatalf("expected test-policy, got %s", pa.Index)
	}
}

func TestUpdate_Refresh(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{}`))
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
		t.Fatal("expected command to fetch policies")
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
	if !strings.Contains(info, "policies") {
		t.Fatalf("expected 'policies' in status info, got %s", info)
	}
}
