# Usage

Quick reference for using `es-cli`. For install & config, see [README](README.md).

## Launching

```bash
es-cli                      # show cluster selector (if multiple)
es-cli --cluster local      # connect directly to a named cluster
es-cli --version            # print version
```

On startup you land in the **index** view.

## Switching views

Press `:` to open the command palette, type a view name, press `Enter`. Tab autocompletes.

| Command                     | View                                            |
| --------------------------- | ----------------------------------------------- |
| `:index` / `:indices`       | List indices (health, docs, size)               |
| `:node` / `:nodes`          | List nodes (CPU, heap, RAM, disk, role)         |
| `:shard` / `:shards`        | List shards (state, docs, store, node)          |
| `:dashboard` / `:dash`      | Cluster overview                                |
| `:ilm` / `:ilm-policy`      | ILM policies                                    |
| `:template` / `:templates`  | Index templates                                 |
| `:discovery`                | Query / log viewer (Kibana Discover style)      |

- `Esc` — pop back to previous view (view stack)
- `q` or `Ctrl+C` — quit
- `?` — full-screen help for current view

## Global keys

| Key     | Action                            |
| ------- | --------------------------------- |
| `:`     | Command palette                   |
| `?`     | Help overlay                      |
| `Esc`   | Back / close overlay              |
| `/`     | Search / filter in current view   |
| `r`     | Refresh current view              |
| `q`     | Quit                              |

## Index view

| Key        | Action                           |
| ---------- | -------------------------------- |
| `Enter`    | Open detail (settings/mappings/aliases) |
| `n`        | New index                        |
| `o`        | Open / close index               |
| `d`        | Delete index (confirm)           |
| `a`        | Toggle hidden (system indices)   |
| `Shift+I`  | Sort by name                     |
| `Shift+H`  | Sort by health                   |
| `Shift+S`  | Sort by size                     |
| `Shift+C`  | Sort by doc count                |

## Node view

| Key | Action                                                   |
| --- | -------------------------------------------------------- |
| `m` | Maintenance menu (cluster routing allocation: primaries / none / reset) |

## Discovery view

1. Select an index
2. Build a query (filter fields, time range)
3. Browse results; press the follow-mode toggle to tail new docs live

## Tips

- Command palette supports aliases and ghost-text autocomplete — type `:ind` + `Tab`.
- When cluster routing allocation is not default, the status bar shows a yellow warning on the right.
- Status bar flash messages auto-clear after ~4s.
- View stack has no depth limit — `Esc` all the way back to where you started.
