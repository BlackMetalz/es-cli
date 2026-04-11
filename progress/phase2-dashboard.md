# Phase 2: Dashboard (Cluster Monitoring)

## Status: Complete

## Features Implemented

### Dashboard View (`:dashboard` / `:dash`)
- 3 styled boxes side-by-side (responsive: vertical below 80 cols)
- Equal height boxes using post-render padding
- Rounded blue borders with padding

**Overview section:**
- Version (e.g., 8.17.0)
- Health — colored green/yellow/red
- Uptime — formatted as "5d 3h 12m"

**Nodes section (title includes count: "Nodes: 1"):**
- Disk Available — e.g., "338.5gb / 376.9gb"
- JVM Heap — e.g., "312.0mb / 512.0mb"

**Indices section (title includes count: "Indices: 131"):**
- Documents — formatted with commas: "100,484"
- Disk Usage — e.g., "187.1mb"
- Primary Shards — e.g., 131
- Replica Shards — e.g., 0

### Keybindings
- `r` — refresh dashboard data
- `?` — help
- `q` — quit

### ES Client
- `dashboard.go` — `DashboardData` struct + `GetDashboardData()` method
- Aggregates 3 API calls:
  - `GET /` — cluster version
  - `GET /_cluster/stats` — indices count/docs/store, nodes count/jvm/fs, shard counts
  - `GET /_nodes/stats/jvm` — uptime from oldest node's `start_time_in_millis`
- Helper functions: `jsonInt()`, `jsonInt64()`, `jsonFloat()` for type-safe JSON navigation
- Uses existing `FormatBytes()` for human-readable sizes

### Layout Details
- Labels: bold white, fixed 16-char width for alignment
- Values: white
- Section titles: bold cyan
- Status bar: "cluster: green" context
- Content padded to fill screen height (status bar sticks to bottom)

### Command Registration
- Registered as `dashboard` with alias `dash` in command router
- Added to `handleCommand()` switch in app.go

## New Files
```
internal/es/dashboard.go (+ dashboard_test.go)
internal/tui/views/dashboard/model.go
internal/tui/views/dashboard/keybindings.go
internal/tui/views/dashboard/model_test.go
```

## Test Coverage
- dashboard ES: 3 tests (happy path, error, zero indices)
- dashboard view: 12 tests (New, Loaded, Error, Refresh, View loading/rendered, HelpGroups, IsInputMode, StatusInfo, SetSize, FormatUptime, FormatNumber)
- 13 packages all passing
