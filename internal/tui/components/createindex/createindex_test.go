package createindex

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew(t *testing.T) {
	m := New()
	if len(m.inputs) != int(fieldCount) {
		t.Fatalf("expected %d inputs, got %d", fieldCount, len(m.inputs))
	}
	if m.focused != fieldName {
		t.Fatal("expected focus on name field")
	}
	if m.inputs[fieldShards].Value() != "1" {
		t.Fatal("expected default shards=1")
	}
	if m.inputs[fieldReplicas].Value() != "1" {
		t.Fatal("expected default replicas=1")
	}
}

func TestUpdate_Tab(t *testing.T) {
	m := New()

	// Tab to shards
	keyMsg := tea.KeyMsg{Type: tea.KeyTab}
	m, _ = m.Update(keyMsg)
	if m.focused != fieldShards {
		t.Fatalf("expected focus on shards, got %d", m.focused)
	}

	// Tab to replicas
	m, _ = m.Update(keyMsg)
	if m.focused != fieldReplicas {
		t.Fatalf("expected focus on replicas, got %d", m.focused)
	}

	// Tab wraps to name
	m, _ = m.Update(keyMsg)
	if m.focused != fieldName {
		t.Fatalf("expected focus on name, got %d", m.focused)
	}
}

func TestUpdate_ShiftTab(t *testing.T) {
	m := New()

	// Shift+tab wraps to replicas
	keyMsg := tea.KeyMsg{Type: tea.KeyShiftTab}
	m, _ = m.Update(keyMsg)
	if m.focused != fieldReplicas {
		t.Fatalf("expected focus on replicas, got %d", m.focused)
	}
}

func TestUpdate_SubmitValid(t *testing.T) {
	m := New()
	m.inputs[fieldName].SetValue("test-index")
	m.inputs[fieldShards].SetValue("3")
	m.inputs[fieldReplicas].SetValue("2")

	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, cmd := m.Update(keyMsg)

	if cmd == nil {
		t.Fatal("expected command on submit")
	}

	msg := cmd()
	submit, ok := msg.(SubmitMsg)
	if !ok {
		t.Fatal("expected SubmitMsg")
	}
	if submit.Name != "test-index" || submit.Shards != 3 || submit.Replicas != 2 {
		t.Fatalf("unexpected submit: %+v", submit)
	}
}

func TestUpdate_SubmitEmptyName(t *testing.T) {
	m := New()
	m.inputs[fieldName].SetValue("")

	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	m, _ = m.Update(keyMsg)

	if m.err == "" {
		t.Fatal("expected error for empty name")
	}
}

func TestUpdate_SubmitInvalidShards(t *testing.T) {
	m := New()
	m.inputs[fieldName].SetValue("test")
	m.inputs[fieldShards].SetValue("abc")

	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	m, _ = m.Update(keyMsg)

	if m.err == "" {
		t.Fatal("expected error for invalid shards")
	}
}

func TestUpdate_SubmitZeroShards(t *testing.T) {
	m := New()
	m.inputs[fieldName].SetValue("test")
	m.inputs[fieldShards].SetValue("0")

	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	m, _ = m.Update(keyMsg)

	if m.err == "" {
		t.Fatal("expected error for zero shards")
	}
}

func TestUpdate_Cancel(t *testing.T) {
	m := New()

	keyMsg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := m.Update(keyMsg)

	if cmd == nil {
		t.Fatal("expected command on cancel")
	}

	msg := cmd()
	if _, ok := msg.(CancelMsg); !ok {
		t.Fatal("expected CancelMsg")
	}
}

func TestView_ContainsFields(t *testing.T) {
	m := New()
	view := m.View()

	if view == "" {
		t.Fatal("expected non-empty view")
	}
}
