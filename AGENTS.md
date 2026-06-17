# SuperLandings Go - Agent Guide

This document helps AI agents work with the SuperLandings Go codebase.

## Project Overview

Go port of SuperLandings with:
- Single binary deployment
- SQLite (metadata) + file system (content)
- Agent-first CLI: all commands output JSON by default
- Dynamic blocks and Go templates
- File system-based versioning

## Tech Stack

- **Runtime**: Go 1.25+
- **Database**: SQLite (modernc.org/sqlite - pure Go)
- **HTTP**: net/http (standard library)
- **CLI**: Cobra
- **Storage**: Hybrid SQLite + file system

## File Size Limits ⚠️ STRICT

| File Type | Max LOC |
|-----------|---------|
| `.md` files | **250** |
| Skills (`.agents/skills/*`) | **300** |
| Go files (`.go`) | **400** |
| JS/HTML/CSS files | **400** |

```bash
wc -l README.md
find . -name "*.md" -exec wc -l {} +
find ./internal -name "*.go" -exec wc -l {} +
```

## Project Structure

```
superlandings-go/
├── cmd/sl-cli/main.go          # CLI entry
├── internal/cli/               # CLI commands (max 400 LOC)
├── internal/config/            # Config (max 400 LOC)
├── internal/daemon/            # Daemon (max 400 LOC)
├── internal/db/                # Database (max 400 LOC each)
├── internal/server/            # HTTP server (max 400 LOC)
├── internal/services/          # Business logic (max 400 LOC each)
├── docs/                       # Documentation (max 250 LOC each)
└── .agents/skills/             # Local skills (max 300 LOC each)
```

## CLI Cheatsheet

```bash
sl-cli site create --name "Site" --slug "site"
sl-cli site list
sl-cli site version create site --version "v1"
sl-cli site version switch site v2
sl-cli site write site v1 "index.html" --content "<html>...</html>"
sl-cli backend start --daemon --port 3099
sl-cli backend start --daemon --port 3099 --no-systemd
sl-cli backend stop
sl-cli backend stop --uninstall
```

## Data Models

### Site
```go
type Site struct {
    ID        string
    Name      string
    Slug      string
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

### SiteVersion
```go
type SiteVersion struct {
    ID        string
    SiteID    string
    Version   string    // "v1", "v2", etc.
    Path      string    // FS path like "sites/foo/v1"
    Comment   string
    Author    string
    IsActive  bool
    CreatedAt time.Time
}
```

### Landing (legacy)
```go
type Landing struct {
    ID, Name, Slug, Type, OrganizationID string
    Content                              string
    Files                                []File
    Domains                              []Domain
    Config                               Config
    CreatedAt, UpdatedAt                 time.Time
}
```

## Templating

**Includes** (no data file):
```html
{{>include "nav.html"}}
<h1>Home</h1>
{{>include "footer.html"}}
```

**Go Templates** (with data file):
```html
<h1>{{.title}}</h1>
{{if .showBanner}}<div>{{.bannerText}}</div>{{end}}
{{range .posts}}<h2>{{.title}}</h2>{{end}}
```
Data: `index.html.data.json`:
```json
{"title":"My Site","showBanner":true,"bannerText":"Welcome!","posts":[{"title":"Post 1"}]}
```

**Processing pipeline:** Includes are resolved first, then if a `.data.json` file exists, the result is rendered with `html/template`.

## Architecture Decisions

| Area | Choice |
|------|--------|
| Database | SQLite (`modernc.org/sqlite`), file: `~/.superlandings/db.sql` |
| Storage | Hybrid: SQLite (metadata) + FS (content) at `~/.superlandings/sites/{slug}/{version}/` |
| Version Control | File system based, instant rollback via DB flag |
| Templates | `{{>include "path"}}` includes + Go `html/template` + `.html.data.json` data files |

## HTTP Server

### Routes
- `/:slug` — Serve site index
- `/:slug/:page` — Serve site page with sub-path routing
- `/` — List all sites
- `/health` — Health check

### Serving Logic
1. Attempt to serve as site (dynamic blocks + Go templates)
2. Fall back to landing
3. Return 404 if not found

### File System Layout

```
~/.superlandings/
├── db.sql
├── sites/{slug}/{version}/{files}
├── sl-cli.pid
└── sl-cli.log
```

### Daemon & Systemd

Systemd unit auto-installed with `--daemon`:
```ini
[Unit]
Description=SuperLandings CLI Daemon
After=network.target
[Service]
Type=simple
User=<user>
WorkingDirectory=<working-dir>
ExecStart=<executable> backend start --port=<port>
Restart=always
[Install]
WantedBy=multi-user.target
```

## Agent Memory (Local Skills) 🧠

Create/update skills under `~/.agents/skills/` for recurring patterns. Keep them **generic** — use placeholders (`<SERVER_IP>`, `<DOMAIN>`) instead of specific IPs or hostnames.

## Common Workflows

### Adding CLI Command
1. File in `internal/cli/` → register in `root.go`
2. Logic in `internal/services/`
3. DB schema in `internal/db/` if needed
4. Build + test

### Adding Service
1. File in `internal/services/`
2. Repository in `internal/db/repository.go` if needed
3. Models in `internal/db/models.go` if needed
4. Wire up in CLI and test

### DB Schema Changes
1. Update models in `internal/db/models.go` (max 400 LOC)
2. Add migration to `internal/db/sqlite.go` (max 400 LOC)
3. Test by deleting `~/.superlandings/db.sql` and restarting
4. Update repository if needed
5. Document in AGENTS.md

## Testing

```bash
go build -o sl-cli ./cmd/sl-cli
go test ./...

# Quick smoke
./sl-cli site create --name "Test" --slug "test"
./sl-cli site version create test --version "v1"
./sl-cli site write test v1 "index.html" --content "<h1>Test</h1>"
./sl-cli backend start --daemon --port 3099
curl http://localhost:3099/test
./sl-cli backend stop
```

## Gotchas

### Traefik Configuration ⚠️ CRITICAL

**NEVER edit Traefik directly.** Use `sl-cli` → `hotify-cli` → improve hotify-cli at `~/ai/hotify-cli`. Manual edits break idempotency.

### Backend API Endpoint Definitions ⚠️ CRITICAL

**Route order matters:** Register `/api/` BEFORE the catch-all `/`. Handlers using `http.StripPrefix("/api", apiMux)` receive paths WITHOUT the `/api/` prefix.

```go
mux.Handle("/api/", http.StripPrefix("/api", apiMux))
mux.HandleFunc("/", handleLanding)  // Catch-all last
// Handler: path is "/sites/x", not "/api/sites/x"
```

**No `defer db.Close()` in server start** — the defer runs when the function returns, closing the DB before any requests arrive. Keep it open for the server lifetime.

### Daemon
- PID: `~/.superlandings/sl-cli.pid`
- Log: `~/.superlandings/sl-cli.log`
- Systemd: `/etc/systemd/system/sl-cli.service`
- Use `--no-systemd` to skip auto-install

### Architecture Decisions to Revisit
- SQLite vs PostgreSQL for high-traffic multi-tenant
- Single binary vs microservices for scaling
- File system vs S3 for cloud-native storage

---

**File size limits are strict. Split before committing. Prefer Go stdlib. Update docs + skills after changes.**
