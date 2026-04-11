package tui

import (
	"net/http"
	"net/http/httptest"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kienlt/es-cli/internal/es"
	"github.com/kienlt/es-cli/internal/tui/components/createindex"
	"github.com/kienlt/es-cli/internal/tui/theme"
	detailview "github.com/kienlt/es-cli/internal/tui/views/detail"
	indexview "github.com/kienlt/es-cli/internal/tui/views/index"
)

func newTestClient() *es.Client {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[]`))
	}))
	return es.NewClient(server.URL, "elastic", "elastic")
}

func TestNewApp(t *testing.T) {
	client := newTestClient()
	app := NewApp(client, "http://localhost:9200")

	if app.header.ClusterURL != "http://localhost:9200" {
		t.Fatal("expected cluster URL set")
	}
	if len(app.viewStack) != 1 {
		t.Fatal("expected 1 view in stack")
	}
	if app.currentView().Name() != "Dashboard" {
		t.Fatalf("expected Dashboard view, got %s", app.currentView().Name())
	}
}

func TestApp_WindowResize(t *testing.T) {
	client := newTestClient()
	app := NewApp(client, "http://localhost:9200")

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	_, _ = app.Update(msg)

	if app.width != 120 {
		t.Fatalf("expected width 120, got %d", app.width)
	}
	if app.height != 40 {
		t.Fatalf("expected height 40, got %d", app.height)
	}
}

func TestApp_OpenCreateIndex(t *testing.T) {
	client := newTestClient()
	app := NewApp(client, "http://localhost:9200")

	// Switch to index view first (default is dashboard)
	app.handleCommand("index")
	app.viewStack[0], _ = app.currentView().Update(indexview.IndicesLoadedMsg{Indices: nil})

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	_, _ = app.Update(keyMsg)

	if app.overlay != overlayCreateIndex {
		t.Fatal("expected create index overlay")
	}
}

func TestApp_CancelCreateIndex(t *testing.T) {
	client := newTestClient()
	app := NewApp(client, "http://localhost:9200")
	app.overlay = overlayCreateIndex
	app.createIndexFm = createindex.New()

	_, _ = app.Update(createindex.CancelMsg{})

	if app.overlay != overlayNone {
		t.Fatal("expected no overlay after cancel")
	}
}

func TestApp_SubmitCreateIndex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPut {
			w.Write([]byte(`{"acknowledged":true}`))
			return
		}
		w.Write([]byte(`[]`))
	}))
	client := es.NewClient(server.URL, "elastic", "elastic")
	app := NewApp(client, server.URL)
	app.overlay = overlayCreateIndex

	_, cmd := app.Update(createindex.SubmitMsg{Name: "new-index", Shards: 1, Replicas: 1})

	if app.overlay != overlayNone {
		t.Fatal("expected overlay closed after submit")
	}
	if cmd == nil {
		t.Fatal("expected command from submit")
	}
}

func TestApp_ConfirmOverlay(t *testing.T) {
	client := newTestClient()
	app := NewApp(client, "http://localhost:9200")
	app.width = 80
	app.height = 24

	// Switch to index view first
	app.handleCommand("index")
	app.currentView().SetSize(80, 24-app.header.Height())

	// Load indices
	app.viewStack[0], _ = app.currentView().Update(indexview.IndicesLoadedMsg{
		Indices: []es.Index{{Name: "test-index", Health: "green", Status: "open"}},
	})

	// Press d to trigger delete
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	_, _ = app.Update(keyMsg)

	if app.overlay != overlayConfirm {
		t.Fatal("expected confirm overlay")
	}
	if app.confirmAction != "delete" {
		t.Fatalf("expected delete action, got %s", app.confirmAction)
	}
	if app.confirmIndex != "test-index" {
		t.Fatalf("expected test-index, got %s", app.confirmIndex)
	}
}

func TestApp_ConfirmCancel(t *testing.T) {
	client := newTestClient()
	app := NewApp(client, "http://localhost:9200")
	app.overlay = overlayConfirm
	app.confirmAction = "delete"
	app.confirmIndex = "test-index"

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	_, _ = app.Update(keyMsg)

	if app.overlay != overlayNone {
		t.Fatal("expected no overlay after cancel")
	}
}

func TestApp_ConfirmAccept(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"acknowledged":true}`))
	}))
	client := es.NewClient(server.URL, "elastic", "elastic")
	app := NewApp(client, server.URL)
	app.overlay = overlayConfirm
	app.confirmAction = "delete"
	app.confirmIndex = "test-index"

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	_, cmd := app.Update(keyMsg)

	if app.overlay != overlayNone {
		t.Fatal("expected no overlay after confirm")
	}
	if cmd == nil {
		t.Fatal("expected command from confirm")
	}
}

func TestApp_ConfirmViewCentered(t *testing.T) {
	client := newTestClient()
	app := NewApp(client, "http://localhost:9200")
	app.width = 80
	app.height = 24
	app.header.Width = 80
	app.overlay = overlayConfirm
	app.confirmAction = "delete"
	app.confirmIndex = "my-index"

	view := app.View()
	if view == "" {
		t.Fatal("expected non-empty view")
	}
}

func TestApp_View(t *testing.T) {
	client := newTestClient()
	app := NewApp(client, "http://localhost:9200")
	app.width = 80
	app.height = 24
	app.header.Width = 80

	view := app.View()
	if view == "" {
		t.Fatal("expected non-empty view")
	}
}

func TestApp_ViewStack(t *testing.T) {
	client := newTestClient()
	app := NewApp(client, "http://localhost:9200")

	if len(app.viewStack) != 1 {
		t.Fatalf("expected 1 view, got %d", len(app.viewStack))
	}
	if app.currentView().Name() != "Dashboard" {
		t.Fatalf("expected Dashboard, got %s", app.currentView().Name())
	}
}

func TestApp_PushPopView(t *testing.T) {
	client := newTestClient()
	app := NewApp(client, "http://localhost:9200")
	app.width = 80
	app.height = 24

	// Switch to index view first
	app.handleCommand("index")

	// Load indices
	app.viewStack[0], _ = app.currentView().Update(indexview.IndicesLoadedMsg{
		Indices: []es.Index{{Name: "test-idx", Health: "green", Status: "open"}},
	})
	app.currentView().SetSize(80, app.viewHeight())

	// Press enter to view detail
	keyMsg := tea.KeyMsg{Type: tea.KeyEnter}
	_, _ = app.Update(keyMsg)

	if len(app.viewStack) != 2 {
		t.Fatalf("expected 2 views after enter, got %d", len(app.viewStack))
	}

	// Pop via GoBackMsg
	_, _ = app.Update(detailview.GoBackMsg{})
	if len(app.viewStack) != 1 {
		t.Fatalf("expected 1 view after back, got %d", len(app.viewStack))
	}
	if app.currentView().Name() != "Indices" {
		t.Fatalf("expected Indices after back, got %s", app.currentView().Name())
	}
}

func TestApp_PopViewMinimum(t *testing.T) {
	client := newTestClient()
	app := NewApp(client, "http://localhost:9200")

	// Pop on single view should not crash
	app.popView()
	if len(app.viewStack) != 1 {
		t.Fatal("expected stack to remain at 1")
	}
}

func TestApp_HelpOverlay(t *testing.T) {
	client := newTestClient()
	app := NewApp(client, "http://localhost:9200")
	app.width = 80
	app.height = 24

	// Load indices first
	app.viewStack[0], _ = app.currentView().Update(indexview.IndicesLoadedMsg{Indices: nil})

	// Press ? to open help
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if app.overlay != overlayHelp {
		t.Fatal("expected help overlay")
	}

	// Press esc to close
	_, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if app.overlay != overlayNone {
		t.Fatal("expected no overlay after esc")
	}
}

func TestApp_FlashMessage(t *testing.T) {
	client := newTestClient()
	app := NewApp(client, "http://localhost:9200")

	cmd := app.setFlash("test message", theme.StatusBarSuccessStyle)
	if app.flashMsg != "test message" {
		t.Fatal("expected flash message set")
	}
	if cmd == nil {
		t.Fatal("expected tick command")
	}

	// Clear flash
	_, _ = app.Update(clearFlashMsg{})
	if app.flashMsg != "" {
		t.Fatal("expected flash cleared")
	}
}
