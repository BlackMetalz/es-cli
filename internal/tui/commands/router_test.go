package commands

import "testing"

func TestRouter_Match(t *testing.T) {
	r := NewRouter()
	r.Register(Command{Name: "index", Description: "List indices"})
	r.Register(Command{Name: "node", Aliases: []string{"nodes"}, Description: "List nodes"})

	if cmd := r.Match("index"); cmd == nil || cmd.Name != "index" {
		t.Fatal("expected match for 'index'")
	}
	if cmd := r.Match("nodes"); cmd == nil || cmd.Name != "node" {
		t.Fatal("expected match for 'nodes' alias")
	}
	if cmd := r.Match("unknown"); cmd != nil {
		t.Fatal("expected nil for unknown command")
	}
}

func TestRouter_Complete(t *testing.T) {
	r := NewRouter()
	r.Register(Command{Name: "index"})
	r.Register(Command{Name: "ilm"})
	r.Register(Command{Name: "node"})

	matches := r.Complete("in")
	if len(matches) != 1 || matches[0].Name != "index" {
		t.Fatalf("expected [index], got %v", matches)
	}

	matches = r.Complete("i")
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches for 'i', got %d", len(matches))
	}

	matches = r.Complete("")
	if len(matches) != 3 {
		t.Fatalf("expected all 3 commands, got %d", len(matches))
	}
}

func TestRouter_Names(t *testing.T) {
	r := NewRouter()
	r.Register(Command{Name: "index"})
	r.Register(Command{Name: "node"})

	names := r.Names()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
}
