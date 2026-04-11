package allocationmenu

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew(t *testing.T) {
	m := New("primaries")
	if m.current != "primaries" {
		t.Fatalf("expected primaries, got %s", m.current)
	}
	if m.cursor != 0 {
		t.Fatal("expected cursor at 0")
	}
}

func TestCursorMovement(t *testing.T) {
	m := New("")

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 1 {
		t.Fatalf("expected cursor 1, got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 2 {
		t.Fatalf("expected cursor 2, got %d", m.cursor)
	}

	// Wrap around
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.cursor != 0 {
		t.Fatalf("expected cursor 0 (wrapped), got %d", m.cursor)
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.cursor != 2 {
		t.Fatalf("expected cursor 2 (wrapped up), got %d", m.cursor)
	}
}

func TestSubmit(t *testing.T) {
	m := New("")
	// Move to "primaries"
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	submit, ok := msg.(SubmitMsg)
	if !ok {
		t.Fatal("expected SubmitMsg")
	}
	if submit.Value != "primaries" {
		t.Fatalf("expected primaries, got %s", submit.Value)
	}
}

func TestSubmit_Reset(t *testing.T) {
	m := New("")
	// First item is "all (reset)" with empty value
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	msg := cmd()
	submit := msg.(SubmitMsg)
	if submit.Value != "" {
		t.Fatalf("expected empty string for reset, got %s", submit.Value)
	}
}

func TestCancel(t *testing.T) {
	m := New("")
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	if _, ok := msg.(CancelMsg); !ok {
		t.Fatal("expected CancelMsg")
	}
}

func TestView(t *testing.T) {
	m := New("primaries")
	view := m.View()
	if !strings.Contains(view, "Cluster Routing Allocation") {
		t.Fatal("expected title")
	}
	if !strings.Contains(view, "primaries") {
		t.Fatal("expected primaries in view")
	}
}
