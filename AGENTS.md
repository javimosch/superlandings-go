# SuperLandings Go - Agent Guide

This document helps AI agents work with the SuperLandings Go codebase.

## Project Overview

Go port of SuperLandings with:
- Single binary deployment
- SQLite (metadata) + file system (content)
- Agent-first CLI: all commands output JSON by default with semantic exit codes
- Dynamic blocks, Go templates, shared assets (`{{asset}}`)
- File system-based versioning, instant rollback
- Schema-driven admin panel (blog editor, raw HTML editor, form editor)
- Domain-aware serving via Host header (including root path)

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
├── cmd/sl-cli/main.go
├── internal/cli/ server/ services/ db/ daemon/ config/
├── docs/ .agents/skills/
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

## Architecture Decisions

| Area | Choice |
|------|--------|
| Database | SQLite (`modernc.org/sqlite`), file: `~/.superlandings/db.sql` |
| Storage | Hybrid: SQLite (metadata) + FS (content) at `~/.superlandings/sites/{slug}/{version}/` + shared `assets/` dir |
| Version Control | File system based, instant rollback via DB flag |
| Templates | `{{>include "path"}}` includes + Go `html/template` + `{{asset "file"}}` resolver |

## HTTP Server

### Routes
- `/:slug` — Serve site index
- `/:slug/:page` — Serve site page with sub-path routing
- `/` — List all sites
- `/health` — Health check

### Serving Logic
Site with Go templates + includes → 404.

### File System Layout

```
~/.superlandings/
├── db.sql
├── sites/{slug}/
│   ├── assets/               # Shared across versions
│   ├── admin-schema.json      # Site-level admin config
│   └── {version}/
│       ├── index.html
│       ├── index.html.data.json
│       ├── layout.html         # Blog post wrapper
│       ├── blog-preview.html   # Blog preview include
│       └── blog/
│           ├── post.md
│           └── post.md.data.json
├── sl-cli.pid
└── sl-cli.log
```

### Daemon

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

- **CLI feature:** file in `internal/cli/` → register in `root.go` → logic in `internal/services/`
- **DB change:** update models → migration in `sqlite.go` → test `rm ~/.superlandings/db.sql`
- **Admin panel:** edit `admin.go` embedded HTML/CSS/JS (single file)
- **Blog:** `blog/*.md` + `.md.data.json` metadata, `layout.html` for styling

## Testing

```bash
go build -o sl-cli ./cmd/sl-cli && go test ./...
# Smoke: create site, write page, start daemon, curl, stop
```

## Gotchas

## Gotchas

- **Route order:** register `/api/` BEFORE catch-all `/`. StripPrefix handlers see paths WITHOUT prefix.
- **No `defer db.Close()`** in server start — closes DB before requests arrive.
- **Template `.Funcs()` before `.Parse()`** — Go requirement.
- **WAL checkpoint** — call `db.CheckpointWAL()` after token/user writes so daemon sees them.
- **Domain root path** — host-based resolution must happen before `handleRoot` for `/`.
- **Admin schema** — `admin-schema.json` at site level, not in version dir.
- **Roles** — see [docs/roles.md](docs/roles.md) for RBAC table (admin/editor/viewer).
- **Toast** — `pointer-events:none` when hidden, `auto` on `.show`.
- **Login field** — `type="text"` (not email) for non-email usernames.
- **Shared assets** — in `sites/{slug}/assets/`, not per-version.
- **Traefik perms** — daemon user needs write to `/etc/traefik/*.yml` + sudo for restart.
- **Hotify-cli config** — same `~/.hotify/config.json` for daemon and infra user.

### Daemon
- PID: `~/.superlandings/sl-cli.pid`, Log: `~/.superlandings/sl-cli.log`
- Systemd: `/etc/systemd/system/sl-cli.service`, use `--no-systemd` to skip

---

**File size limits are strict. Split before committing. Prefer Go stdlib. Update docs + skills after changes.**
