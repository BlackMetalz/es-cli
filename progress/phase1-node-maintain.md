# Phase 1: Command Palette + Node View + Shard View + Node Maintenance

## Status: Complete

## Features Implemented

### Command Palette
- **`:` key** opens command input at the bottom (replaces status bar)
- **Autocomplete** — typing `:no` shows ghost text `de` in faint cyan, Tab completes
- **Commands**: `:index` (indices), `:node`/`:nodes`, `:shard`/`:shards`
- **Enter** executes matched command (switches view), **Esc** cancels
- Enter also works with partial + ghost text (e.g., type "no" + Enter → matches "node")
- Command router: `internal/tui/commands/router.go` with Register/Match/Complete

### Node View (`:node`)
- Table: name, ip, heap%, ram%, cpu, load_1m, load_5m, load_15m, role, master, disk%
- **Sort**: Shift+N (name), Shift+C (cpu), Shift+H (heap), Shift+R (ram), Shift+D (disk) — toggle ASC/DESC
- **Search**: `/` to filter by node name
- **Colorized**: values >80% in red, >60% in yellow, master `*` in green
- **Maintenance**: `m` key opens allocation menu overlay

### Shard View (`:shard`)
- Table: index, shard, prirep, state, docs, store, ip, node
- **Sort**: Shift+I (index), Shift+S (shard), Shift+T (state), Shift+N (node), Shift+D (docs), Shift+O (store)
- **Search**: `/` to filter by index name or node name
- **Colorized states**: STARTED=green, RELOCATING/INITIALIZING=yellow, UNASSIGNED=red
- **Hide system indices**: `a` toggles, default hides `.` prefixed, shows hidden count

### Node Maintenance (Allocation Menu)
- Press `m` in node view → fetches current allocation setting → shows modal overlay
- 3 options: `all (reset)`, `primaries`, `none`
- j/k or up/down navigates, Enter selects, Esc cancels
- Calls `PUT /_cluster/settings` with transient `cluster.routing.allocation.enable`
- Flash message on success: "Allocation set to primaries"
- **Status bar indicator**: shows `alloc: primaries` in yellow when not default

### ES Client Additions
- `node.go` — `Node` struct + `ListNodes()` via `GET /_cat/nodes?format=json`
- `shard.go` — `Shard` struct + `ListShards()` via `GET /_cat/shards?format=json`
- `cluster.go` — `GetAllocationSetting()` + `SetAllocationSetting()` via `GET/PUT /_cluster/settings`
- `helpers.go` — exported `JsonStr()` helper (was `jsonStr` in index.go)

### App.go Changes
- Router registration: index, node, shard commands with aliases
- `handleCommand()` switch: creates appropriate view + calls Init
- `switchView()` — replaces entire viewStack (top-level navigation)
- `handlePendingAction()` — handles "set_allocation" type from node view
- Allocation overlay routing + flash messages

## New Files
```
internal/es/node.go, shard.go, cluster.go, helpers.go (+ tests)
internal/tui/components/cmdpalette/cmdpalette.go (+ test)
internal/tui/components/allocationmenu/allocationmenu.go (+ test)
internal/tui/views/node/model.go, keybindings.go, sort.go, render.go, filter.go (+ test)
internal/tui/views/shard/model.go, keybindings.go, sort.go, render.go, filter.go (+ test)
```

## Test Coverage
- 13 packages all passing
- cmdpalette: 95.2%
- allocationmenu: 96.6%
- ES client: 86.4%
- Node/shard views: tested (New, Loaded, Sort, Refresh, Error, HelpGroups, SetSize, IsInputMode, StatusInfo)
