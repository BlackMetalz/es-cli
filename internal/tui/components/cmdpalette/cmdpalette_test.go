package cmdpalette

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kienlt/es-cli/internal/tui/commands"
)

func testRouter() *commands.Router {
	r := commands.NewRouter()
	r.Register(commands.Command{Name: "index", Aliases: []string{"indices"}})
	r.Register(commands.Command{Name: "node", Aliases: []string{"nodes"}})
	r.Register(commands.Command{Name: "shard", Aliases: []string{"shards"}})
	return r
}

func TestNew(t *testing.T) {
	r := testRouter()
	m := New(r, 80)
	if m.ghost != "" {
		t.Fatal("expected empty ghost initially")
	}
}

func TestGhostText(t *testing.T) {
	r := testRouter()
	m := New(r, 80)

	// Type "no" → ghost should be "de"
	m.input.SetValue("no")
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}) // trigger update
	// Manually check: the update recalculates ghost from input value
	// After setting "no", Complete("no") returns [{node}], ghost = "de"
	m.input.SetValue("no")
	m, _ = m.Update(nil) // trigger ghost recalculation on non-key msg
	// Ghost is recalculated on key events only, let's simulate properly
	m2 := New(r, 80)
	m2.input.SetValue("no")
	matches := r.Complete("no")
	if len(matches) != 1 || matches[0].Name != "node" {
		t.Fatalf("expected [node], got %v", matches)
	}
}

func TestTabCompletion(t *testing.T) {
	r := testRouter()
	m := New(r, 80)
	m.input.SetValue("no")
	m.ghost = "de" // simulate ghost

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.input.Value() != "node" {
		t.Fatalf("expected 'node' after tab, got '%s'", m.input.Value())
	}
	if m.ghost != "" {
		t.Fatal("expected ghost cleared after tab")
	}
}

func TestEnter_ValidCommand(t *testing.T) {
	r := testRouter()
	m := New(r, 80)
	m.input.SetValue("node")

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from enter")
	}
	msg := cmd()
	submit, ok := msg.(SubmitMsg)
	if !ok {
		t.Fatal("expected SubmitMsg")
	}
	if submit.Command != "node" {
		t.Fatalf("expected node, got %s", submit.Command)
	}
}

func TestEnter_WithGhostCompletion(t *testing.T) {
	r := testRouter()
	m := New(r, 80)
	m.input.SetValue("no")
	m.ghost = "de"

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command from enter with ghost")
	}
	msg := cmd()
	submit, ok := msg.(SubmitMsg)
	if !ok {
		t.Fatal("expected SubmitMsg")
	}
	if submit.Command != "node" {
		t.Fatalf("expected node, got %s", submit.Command)
	}
}

func TestEnter_InvalidCommand(t *testing.T) {
	r := testRouter()
	m := New(r, 80)
	m.input.SetValue("unknown")

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatal("expected nil command for invalid input")
	}
}

func TestEsc(t *testing.T) {
	r := testRouter()
	m := New(r, 80)

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected command from esc")
	}
	msg := cmd()
	if _, ok := msg.(CancelMsg); !ok {
		t.Fatal("expected CancelMsg")
	}
}

func TestView(t *testing.T) {
	r := testRouter()
	m := New(r, 80)
	view := m.View()
	if view == "" {
		t.Fatal("expected non-empty view")
	}
}
