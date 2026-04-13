# es-cli

A K9s-style terminal UI for managing Elasticsearch clusters. Built in Go using the Bubble Tea TUI framework.

## Features

- Browse indices, nodes, shards with live stats
- Index CRUD (create, open, close, delete)
- ILM policy management
- Index template management
- Discovery view (Kibana Discover style) with query builder + follow mode
- Multi-cluster support with selection UI
- Cluster routing allocation control (maintenance mode)
- Command palette (`:`) for view switching

## Install

### Download prebuilt binary

Grab the latest release from the [Releases page](https://github.com/kienlt/es-cli/releases).

**Linux (amd64):**
```bash
curl -L -o es-cli https://github.com/kienlt/es-cli/releases/latest/download/es-cli-linux-amd64
chmod +x es-cli
sudo mv es-cli /usr/local/bin/
```

**Linux (arm64):**
```bash
curl -L -o es-cli https://github.com/kienlt/es-cli/releases/latest/download/es-cli-linux-arm64
chmod +x es-cli
sudo mv es-cli /usr/local/bin/
```

**macOS (Apple Silicon / M1/M2/M3):**
```bash
curl -L -o es-cli https://github.com/kienlt/es-cli/releases/latest/download/es-cli-darwin-arm64
chmod +x es-cli
xattr -d com.apple.quarantine es-cli  # remove Gatekeeper quarantine
sudo mv es-cli /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -L -o es-cli https://github.com/kienlt/es-cli/releases/latest/download/es-cli-darwin-amd64
chmod +x es-cli
xattr -d com.apple.quarantine es-cli
sudo mv es-cli /usr/local/bin/
```

> **Note on macOS:** the binaries are not code-signed/notarized. When you download, macOS marks them with the `com.apple.quarantine` attribute and Gatekeeper will block execution with `"cannot be opened because the developer cannot be verified"`. Strip the quarantine attribute with `xattr -d com.apple.quarantine es-cli` (as shown above) before running. Alternatively right-click the binary in Finder → Open → Open anyway.

### Build from source

```bash
git clone https://github.com/kienlt/es-cli
cd es-cli
make build
./es-cli
```

Requires Go 1.26+.

## Usage

See [USAGE.md](USAGE.md) for a quick reference of views, keybindings, and the command palette.

## Configuration

Create `~/.es-cli.auth` with your cluster credentials:

```json
{
  "local": {
    "username": "elastic",
    "password": "elastic",
    "url": "http://localhost:9200"
  },
  "staging": {
    "username": "admin",
    "password": "secret",
    "url": "https://staging.es.example.com:9200"
  }
}
```

- Multiple clusters → selection UI on startup
- Single cluster → auto-connects
- Use `--cluster <name>` to skip the selector

## Local development

```bash
make dev-init         # start local ES + Kibana via Docker
make dev-auth         # create sample ~/.es-cli.auth
make run              # build and run
make dev-destroy      # tear down
```

## Release

Tag a commit with `v*` to trigger a release build:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The workflow builds Linux (amd64/arm64) + macOS (amd64/arm64) binaries and publishes a GitHub Release.
