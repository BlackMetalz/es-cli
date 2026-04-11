package createtemplate

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNew(t *testing.T) {
	m := New([]string{"delete-after-30d", "logs-policy", "metrics-policy"}, nil)
	if m.focused != fieldName {
		t.Fatalf("expected focus on name, got %d", m.focused)
	}
}

func TestTabNavigation(t *testing.T) {
	m := New([]string{"delete-after-30d", "logs-policy", "metrics-policy"}, nil)
	keyMsg := tea.KeyMsg{Type: tea.KeyTab}

	m, _ = m.Update(keyMsg)
	if m.focused != fieldPatterns {
		t.Fatalf("expected patterns, got %d", m.focused)
	}

	m, _ = m.Update(keyMsg)
	if m.focused != fieldShards {
		t.Fatalf("expected shards, got %d", m.focused)
	}

	m, _ = m.Update(keyMsg)
	if m.focused != fieldReplicas {
		t.Fatalf("expected replicas, got %d", m.focused)
	}

	m, _ = m.Update(keyMsg)
	if m.focused != fieldILMPolicy {
		t.Fatalf("expected ILM policy, got %d", m.focused)
	}

	m, _ = m.Update(keyMsg)
	if m.focused != fieldName {
		t.Fatalf("expected name (wrap), got %d", m.focused)
	}
}

func TestSubmit_Valid(t *testing.T) {
	m := New([]string{"delete-after-30d", "logs-policy", "metrics-policy"}, nil)
	m.inputs[fieldName].SetValue("my-template")
	m.inputs[fieldPatterns].SetValue("logs-*,metrics-*")
	m.inputs[fieldShards].SetValue("1")
	m.inputs[fieldReplicas].SetValue("1")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command")
	}
	msg := cmd()
	submit := msg.(SubmitMsg)
	if submit.Name != "my-template" {
		t.Fatalf("expected my-template, got %s", submit.Name)
	}
	if !strings.Contains(submit.Body, `"priority": 100`) {
		t.Fatal("expected priority 100 in body")
	}
}

func TestSubmit_EmptyName(t *testing.T) {
	m := New([]string{"delete-after-30d", "logs-policy", "metrics-policy"}, nil)
	m.inputs[fieldPatterns].SetValue("logs-*")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.err == "" {
		t.Fatal("expected error")
	}
}

func TestSubmit_EmptyPatterns(t *testing.T) {
	m := New([]string{"delete-after-30d", "logs-policy", "metrics-policy"}, nil)
	m.inputs[fieldName].SetValue("test")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.err == "" {
		t.Fatal("expected error")
	}
}

func TestCancel(t *testing.T) {
	m := New([]string{"delete-after-30d", "logs-policy", "metrics-policy"}, nil)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	msg := cmd()
	if _, ok := msg.(CancelMsg); !ok {
		t.Fatal("expected CancelMsg")
	}
}

func TestBuildJSON(t *testing.T) {
	m := New([]string{"delete-after-30d", "logs-policy", "metrics-policy"}, nil)
	m.inputs[fieldPatterns].SetValue("logs-*")
	m.inputs[fieldShards].SetValue("2")
	m.inputs[fieldReplicas].SetValue("1")

	json, errMsg := m.buildJSON()
	if errMsg != "" {
		t.Fatalf("unexpected error: %s", errMsg)
	}
	if !strings.Contains(json, `"priority": 100`) {
		t.Fatal("expected priority 100")
	}
	if !strings.Contains(json, `"logs-*"`) {
		t.Fatal("expected logs-* pattern")
	}
}
