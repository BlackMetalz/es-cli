# Phase 4: Multi-Cluster Support

## Status: Complete

## Features Implemented

### New Auth Format (`~/.es-cli.auth`)
- Multi-cluster JSON format: `{"name": {"username":"...", "password":"...", "url":"..."}}`
- Each key is a cluster name, value has username, password, url
- Validates all fields non-empty per cluster
- Sorted by name for stable ordering
- **Old format detection**: if old `{"user":"pass"}` format detected, shows clear migration error with example of new format

### Cluster Selection UI
- Full-screen selection on startup when 2+ clusters configured
- j/k or up/down to navigate, Enter to connect, q/Esc to quit
- Shows cluster name + URL for each entry
- Standalone Bubble Tea model (runs before main app)
- **Single cluster**: auto-connects, skips selection
- **`--cluster` flag**: `./es-cli --cluster local` connects directly without selection
- **Unknown cluster**: shows error with list of available cluster names

### Header Changes
- Shows **config cluster name** (from auth file) instead of ES cluster_name
- e.g., "local" instead of "docker-cluster"
- ClusterName field was already in header Model, now set from config during NewApp

### Docker: Second ES Instance
- Added `es02` service to docker-compose.yml
- Port 9201 (maps to internal 9200), same creds (elastic:elastic)
- Same config as es01: single-node, 512MB heap, 2GB memory limit
- `make dev-init` now waits for both clusters to be healthy

### Makefile Updates
- `make dev-init`: starts all containers, waits for both es01 + es02
- `make dev-auth`: creates sample `~/.es-cli.auth` with local (9200) + local-2 (9201)

### App Signature Change
- `NewApp(client, clusterURL, clusterName)` — accepts cluster name from config
- `fetchClusterInfo` no longer overwrites ClusterName (keeps config name)

## Files Modified
```
internal/auth/auth.go           — rewritten: ClusterConfig type, multi-cluster LoadAuth
internal/auth/auth_test.go      — rewritten: 9 tests for new format
cmd/es-cli/main.go              — --cluster flag, cluster selection flow
internal/tui/app.go             — NewApp accepts clusterName param
docker-compose.yml              — added es02 service
Makefile                        — updated dev-init, added dev-auth
```

## New Files
```
internal/tui/clusterselect/clusterselect.go  — cluster selection TUI model
```

## Test Coverage
- Auth: 9 tests (single/multi cluster, old format, missing fields, empty, not found, invalid JSON)
- All 17 packages passing
