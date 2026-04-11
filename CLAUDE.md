# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is this?

A K9s-style terminal UI for managing Elasticsearch clusters. Built in Go using the Bubble Tea TUI framework.

## Common Commands

```bash
make build          # Build binary to ./es-cli
make test           # Run all tests with race detection and coverage
make run            # Build and run
make dev-init       # Start local ES + Kibana via Docker Compose (sets kibana_system password)
make dev-stop       # Stop containers
make dev-destroy    # Destroy containers and volumes

# Generate test indices
make generate-indices          # Create 100 demo indices (default)
make generate-indices NUM=500  # Create 500 demo indices

# Run a single test
go test ./internal/es/ -run TestFormatBytes -v
```

## Auth Setup

The app reads credentials from `~/.es-cli.auth` as `{"username":"password"}`. ES URL defaults to `http://localhost:9200`, overridable via `ES_URL` env var.

## Architecture

### TUI layer (`internal/tui/`)

Uses Bubble Tea's Elm architecture (Model → Update → View):

- **`App`** (`app.go`) — root model. Manages view stack, header, overlays (create-index, confirm, help), status bar with flash messages. Routes messages to overlays or the active view. No type assertions on views — uses the `View` interface exclusively.
- **`help.go`** — full-screen help overlay renderer (K9s-style grouped columns).
- **`overlay.go`** — utility for compositing overlay text onto background at x,y coordinates.

### Views (`internal/tui/views/`)

- **`View` interface** (`view.go`) — contract for all swappable views:
  - `Init`, `Update`, `View`, `Name`, `SetSize` — standard Bubble Tea lifecycle
  - `HelpGroups()` — returns grouped keybindings for header and help screen
  - `IsInputMode()` — tells App when the view is capturing text input (search, etc.)
  - `PopPendingAction()` — returns and clears any pending action (delete, close, view_detail, etc.)
  - `StatusInfo()` — context text for the status bar
- **`views/index/`** — index list view. Split into:
  - `model.go` — struct, constructor, Update, View, interface methods
  - `keybindings.go` — key definitions
  - `sort.go` — sorting logic with natural sort (demo-2 < demo-10), health ranking
  - `filter.go` — system index hiding, search filter
  - `render.go` — table rendering, column widths, health colorization, selected row highlight
- **`views/detail/`** — index detail view with 3 tabs (Settings, Mappings, Aliases). JSON pretty-print with syntax coloring. Scrollable viewport.

### Command Router (`internal/tui/commands/`)

- **`router.go`** — command registry with match and autocomplete for Phase 1 (`:index`, `:node`, `:shard`).

### ES client (`internal/es/`)

Thin HTTP wrapper around ES REST API:
- `client.go` — HTTP client with basic auth, `GetClusterInfo()` (name, version, health)
- `index.go` — index CRUD (list, create, open, close, delete), size parsing/formatting
- `detail.go` — fetch index settings, mappings, aliases

### Other packages

- **`internal/auth/`** — loads credentials from JSON file on disk.
- **`internal/tui/header/`** — K9s-style header with cluster info (left) + grouped keybindings (right). Dynamic height based on help groups.
- **`internal/tui/theme/`** — centralized color scheme and lipgloss styles.
- **`internal/tui/components/createindex/`** — modal form for creating a new index.

## Key Patterns

- **View stack** — `App` holds a `viewStack []views.View`. `pushView()` navigates forward (e.g., index list → detail), `popView()` navigates back (Esc). Supports unlimited depth.
- **Pending actions** — views set `PendingAction{Type: "delete", Index: "x"}` which App picks up after routing Update. String-based types are extensible without enum changes.
- **No type assertions** — App interacts with views only through the `View` interface. No `(*indexview.Model)` casts.
- **ES operations** run as Bubble Tea commands (closures returning `tea.Msg`), keeping the UI non-blocking.
- **Post-process rendering** — health column colors and selected row highlight are applied after the bubbles table renders, because the table's `runewidth.Truncate` is not ANSI-aware.
- **Flash messages** — `App.setFlash()` shows a timed status bar message (e.g., "Index 'x' deleted successfully") that auto-clears after 4 seconds.
