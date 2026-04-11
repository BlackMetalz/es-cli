# Post-MVP Enhancements

## Status: Complete

## Features Implemented

### UI Improvements
- **K9s-style header redesign** — left: cluster info (URL, Cluster, Health, User, ES Rev), right: grouped keybindings in columns
- **Health coloring** — cluster health in header (green/yellow/red), index health in table via post-processing (bubbles table doesn't support per-cell ANSI)
- **Selected row highlight** — index name highlighted on selected row, works with scroll
- **Sort toggle** — Shift+I/S/C/H toggles ASC ↑ / DESC ↓, with arrow indicator
- **Natural sort** — demo-index-num-2 sorts before demo-index-num-10
- **Search/filter** — `/` opens search input, live filtering, Enter confirms, Esc clears
- **Hide system indices** — default hides `.internal.*` indices, `a` toggles, shows "12 hidden" count
- **Full-screen help** — `?` opens K9s-style help overlay with grouped keybindings, Esc closes
- **Status bar** — bottom bar with view name, context info, flash messages (auto-clear 4s)
- **Index detail view** — Enter opens 3-tab view (Settings/Mappings/Aliases) with syntax-colored JSON, Tab/Shift+Tab switches tabs, Esc goes back

### Keybinding Changes
- `d` — delete index (was Shift+D)
- `o` — smart open/close toggle (detects current status)
- `a` — toggle hidden indices
- `enter` — view index detail
- `Shift+H` — sort by health

### ES Client Additions
- `GetClusterInfo()` — cluster name, version, health from `GET /` + `GET /_cluster/health`
- `GetIndexDetail()` — settings, mappings, aliases
- `OpenIndex()` — `POST /{index}/_open`

### Architecture Refactor
- **View interface extended** — added `IsInputMode()`, `PopPendingAction()`, `StatusInfo()` to eliminate type assertions in app.go
- **Pending actions** — string-based `PendingAction{Type, Index}` replaces old enum
- **View stack** — `viewStack []views.View` replaces single `previousView`, supports unlimited depth
- **File splitting** — index view split into model.go (408), sort.go (102), render.go (135), filter.go (39), keybindings.go (76)
- **App.go split** — help.go (100), overlay.go (55) extracted
- **Command router** — `internal/tui/commands/router.go` with Register/Match/Complete for Phase 1 readiness

### Scripts
- `scripts/generate-indices.sh` — generate N demo indices with `make generate-indices NUM=100`
- `make dev-init` — auto-sets kibana_system password

## Test Coverage
- Total: ~63.5% (up from 48.9%)
- ES client: 86.2%
- Header: 93.0%
- Commands: 88.9%
- CreateIndex: 88.5%
- Detail view: 72.5%
