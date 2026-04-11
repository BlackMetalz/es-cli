package createilm

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew(t *testing.T) {
	m := New()
	if m.editing {
		t.Fatal("expected not editing")
	}
}

func TestNewEdit(t *testing.T) {
	m := NewEdit("my-policy", "90d")
	if !m.editing {
		t.Fatal("expected editing")
	}
	if m.inputs[fieldName].Value() != "my-policy" {
		t.Fatalf("expected my-policy, got %s", m.inputs[fieldName].Value())
	}
	if m.inputs[fieldDeleteAfter].Value() != "90d" {
		t.Fatalf("expected 90d, got %s", m.inputs[fieldDeleteAfter].Value())
	}
}

func TestTabNavigation(t *testing.T) {
	m := New()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.focused != fieldDeleteAfter {
		t.Fatalf("expected delete after, got %d", m.focused)
	}
}

func TestSubmit_Valid(t *testing.T) {
	m := New()
	m.inputs[fieldName].SetValue("test-policy")
	m.inputs[fieldDeleteAfter].SetValue("30d")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	submit := msg.(SubmitMsg)
	if submit.Name != "test-policy" {
		t.Fatalf("expected test-policy, got %s", submit.Name)
	}
}

func TestSubmit_EmptyName(t *testing.T) {
	m := New()
	m.inputs[fieldDeleteAfter].SetValue("30d")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.err == "" {
		t.Fatal("expected error")
	}
}

func TestCancel(t *testing.T) {
	m := New()
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	msg := cmd()
	if _, ok := msg.(CancelMsg); !ok {
		t.Fatal("expected CancelMsg")
	}
}

func TestBuildJSON(t *testing.T) {
	m := New()
	m.inputs[fieldDeleteAfter].SetValue("90d")
	j := m.BuildJSON()
	if !strings.Contains(j, `"90d"`) {
		t.Fatal("expected 90d in JSON")
	}
}

func TestView(t *testing.T) {
	m := New()
	if !strings.Contains(m.View(), "Create ILM") {
		t.Fatal("expected title")
	}
}
