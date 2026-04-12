# Phase 5: Discovery (Query Log Viewer)

## Status: Complete

## Features Implemented

### Discovery View (`:discovery`)
- **Index Selection**: list all user indices (no system), j/k navigate, / to filter, r to refresh, Enter to select
- **Auto-detect columns**: fetches mapping via `GET /{index}/_mapping`, picks @timestamp + first keyword/text fields
- **Column picker**: `c` opens overlay, space toggles fields on/off, Enter confirms
- **Local timestamp**: @timestamp converted from UTC to local time for readability

### Query Builder (`/`)
- Visual popup for building queries without knowing ES syntax
- `n` or `+` to add filter → select field (type to fuzzy filter, j/k navigate) → enter value
- Multiple filters with AND/OR operators
- `a` toggles AND/OR between filters (only when 2+ filters)
- `d` deletes selected filter
- `Enter` applies — generates properly parenthesized ES query_string
- OR groups wrapped in parentheses: `level:INFO AND (service:worker OR service:api)`
- Existing query parsed back into filters when reopening builder

### Raw Query Fallback (`!`)
- Opens raw text input for advanced ES query_string syntax
- For users who know Lucene syntax and need full control

### Follow Mode (`f`)
- Cycles through: OFF → 1s → 2s → 5s → 10s → OFF
- Re-fetches latest 100 docs each interval (no accumulation, prevents memory issues)
- Status shows `[FOLLOWING: 2s]` (green) or `[PAUSED]`
- Works with active query filters

### ES Client Additions (`es/search.go`)
- `GetFieldMapping(index)` — fetches and flattens nested field mappings from `GET /{index}/_mapping`
- `SearchDocs(index, query, fields, size, searchAfter)` — executes `POST /{index}/_search` with query_string, sort by @timestamp desc
- `GetIndexNames()` — returns non-system index names sorted

### Stream Data Script
- `make stream-data` — continuously inserts log-like data into a specified index
- `make stream-data INDEX=my-logs INTERVAL=500` — custom index + interval
- Generates realistic log fields: @timestamp, level (weighted), service, message, host, status_code, duration_ms
- 10 docs per batch, Ctrl+C to stop

### Bug Fixes
- Command palette ghost text now works with aliases (e.g., `:dis` → ghost `covery`)
- Ghost text color changed to cyan (visible on dark backgrounds)
- Table `SetRows(nil)` before `SetColumns` to prevent panic when column count changes
- Follow mode: replaced accumulating hits with full re-fetch (prevents blank table at 400+ docs)

## Files Created
```
internal/es/search.go + search_test.go
internal/tui/views/query/
  model.go            — main Model, messages, constructor, Update
  update.go           — key handlers for all states
  view.go             — rendering: index selector, query view, column picker
  helpers.go          — getField, defaultColumns, rebuildTable, parseQueryToFilters
  querybuilder.go     — query builder popup: filter list, field selector, value input
  keybindings.go      — /, !, f, c, r, enter, esc, ?, q
  model_test.go       — 13 tests
scripts/stream-data.sh
```

## Files Modified
```
internal/tui/app.go                    — register :discovery command
internal/tui/components/cmdpalette/    — ghost text alias fix
Makefile                               — add stream-data target
```

## Keybindings (Query View)
| Key | Action |
|-----|--------|
| `/` | Open query builder |
| `!` | Open raw query input |
| `f` | Cycle follow interval (OFF/1s/2s/5s/10s) |
| `c` | Column picker |
| `r` | Refresh |
| `enter` | Expand document (JSON) |
| `esc` | Back to index selection |
| `?` | Help |
| `q` | Quit |

## Test Coverage
- ES search: 5 tests (GetFieldMapping, SearchDocs, MatchAll, SearchAfter, GetIndexNames)
- Query view: 13 tests (New, IndicesLoaded, SelectIndex, MappingLoaded, SearchResult, FollowToggle, ErrorMsg, View, HelpGroups, IsInputMode, StatusInfo, GetField, DefaultColumns)
- All 18 packages passing
