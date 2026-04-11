package shard

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

	if m.Name() != "Shards" {
		t.Fatalf("expected Shards, got %s", m.Name())
	}
	if !m.loading {
		t.Fatal("expected loading=true initially")
	}
}

func TestUpdate_ShardsLoaded(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)

	shards := []es.Shard{
		{Index: "b-index", ShardN: "0", PriRep: "p", State: "STARTED", Docs: "100", Store: "4.5kb", IP: "10.0.0.1", Node: "node-1"},
		{Index: "a-index", ShardN: "1", PriRep: "r", State: "STARTED", Docs: "5000", Store: "1.2mb", IP: "10.0.0.2", Node: "node-2"},
	}

	v, _ := m.Update(ShardsLoadedMsg{Shards: shards})
	updated := v.(*Model)

	if updated.loading {
		t.Fatal("expected loading=false after shards loaded")
	}
	if updated.err != nil {
		t.Fatalf("unexpected error: %v", updated.err)
	}
	if len(updated.shards) != 2 {
		t.Fatalf("expected 2 shards, got %d", len(updated.shards))
	}
	if updated.shards[0].Index != "a-index" {
		t.Fatalf("expected a-index first (sorted by index ASC), got %s", updated.shards[0].Index)
	}
}

func TestUpdate_SortByIndex(t *testing.T) {
	client := newTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	m := New(client)
	m.loading = false
	m.shards = []es.Shard{
		{Index: "charlie", ShardN: "0"},
		{Index: "alpha", ShardN: "0"},
		{Index: "bravo", ShardN: "0"},
	}

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'I'}}
	v, _ := m.Update(keyMsg)
	updated := v.(*Model)

	if updated.sortField != SortByIndex {
		t.Fatal("expected sort by index")
	}
	// Default sort is already index ASC, pressing I toggles to DESC
	if updated.filtered[0].Index != "charlie" {
		t.Fatalf("expected charlie first (DESC), got %s", updated.filtered[0].Index)
	}

	v, _ = updated.Update(keyMsg)
	updated = v.(*Model)
	if updated.filtered[0].Index != "alpha" {
		t.Fatalf("expected alpha first (ASC), got %s", updated.filtered[0].Index)
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
		t.Fatal("expected command to fetch shards")
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

	if len(groups) != 2 {
		t.Fatalf("expected 2 help groups, got %d", len(groups))
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
	if !strings.Contains(info, "hiding system shards") {
		t.Fatalf("expected 'hiding system shards', got %s", info)
	}
}
