package header

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/kienlt/es-cli/internal/tui/views"
)

func TestNew(t *testing.T) {
	h := New("http://localhost:9200")
	if h.ClusterURL != "http://localhost:9200" {
		t.Fatalf("expected http://localhost:9200, got %s", h.ClusterURL)
	}
	if h.ViewName != "Indices" {
		t.Fatalf("expected Indices, got %s", h.ViewName)
	}
}

func TestView_ZeroWidth(t *testing.T) {
	h := New("http://localhost:9200")
	h.Width = 0
	if h.View() != "" {
		t.Fatal("expected empty string for zero width")
	}
}

func TestView_WithHelpGroups(t *testing.T) {
	h := New("http://localhost:9200")
	h.Width = 120
	h.ClusterName = "my-cluster"
	h.ESVersion = "8.12.0"
	h.User = "elastic"
	h.HelpGroups = []views.HelpGroup{
		{
			Title: "Sort",
			Bindings: []key.Binding{
				key.NewBinding(key.WithKeys("I"), key.WithHelp("Shift+I", "sort by index")),
				key.NewBinding(key.WithKeys("S"), key.WithHelp("Shift+S", "sort by size")),
			},
		},
		{
			Title: "General",
			Bindings: []key.Binding{
				key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
			},
		},
	}

	view := h.View()

	if !strings.Contains(view, "localhost:9200") {
		t.Fatal("expected cluster URL in header")
	}
	if !strings.Contains(view, "my-cluster") {
		t.Fatal("expected cluster name in header")
	}
	if !strings.Contains(view, "Sort") {
		t.Fatal("expected Sort group title")
	}
	if !strings.Contains(view, "General") {
		t.Fatal("expected General group title")
	}
	if !strings.Contains(view, "sort by index") {
		t.Fatal("expected keybinding description")
	}
}

func TestView_ContainsSeparator(t *testing.T) {
	h := New("http://localhost:9200")
	h.Width = 40
	view := h.View()
	if !strings.Contains(view, "─") {
		t.Fatal("expected separator line in header")
	}
}

func TestView_DefaultsToNA(t *testing.T) {
	h := New("http://localhost:9200")
	h.Width = 80
	view := h.View()

	if !strings.Contains(view, "n/a") {
		t.Fatal("expected n/a for missing cluster info")
	}
}

func TestHeight(t *testing.T) {
	h := New("http://localhost:9200")
	h.HelpGroups = []views.HelpGroup{
		{
			Title: "Index",
			Bindings: []key.Binding{
				key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new")),
				key.NewBinding(key.WithKeys("O"), key.WithHelp("O", "close")),
				key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "delete")),
				key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
			},
		},
	}

	// 4 bindings + 1 title + 1 separator = 6
	if h.Height() != 6 {
		t.Fatalf("expected 6, got %d", h.Height())
	}
}
