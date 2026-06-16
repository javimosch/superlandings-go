<p align="center">
  <img src="https://img.shields.io/badge/version-1.0.0-blue" alt="Version">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
  <img src="https://img.shields.io/badge/go-1.25+-orange" alt="Go">
</p>

<h1 align="center">SuperLandings Go 🚀</h1>

<p align="center">
  <b>Go port of SuperLandings landing page management system.</b><br>
  <b>Agent-first, single binary, no dependencies.</b>
</p>

> Think: "SuperLandings, but in Go with less moving parts and agent-friendly CLI"

## ⚡ TL;DR

> Single-binary landing page management with versioning, dynamic blocks, Go templates, and SQLite persistence.

```bash
# Create a site with versioning
sl-cli site create --name "My Site" --slug "my-site"
sl-cli site version create my-site --version "v1"

# Option 1: Simple includes
sl-cli site write my-site v1 "index.html" --content '{{>include "nav.html"}}<h1>Home</h1>{{>include "footer.html"}}'

# Option 2: Go templates with variables
sl-cli site write my-site v1 "index.html" --content '<h1>{{.title}}</h1>{{if .showBanner}}<div>{{.bannerText}}</div>{{end}}'
# Create index.html.data.json: {"title":"My Site","showBanner":true,"bannerText":"Welcome!"}

# Start daemon
sl-cli backend start --daemon --port 3099
curl http://localhost:3099/my-site/
```

👉 **Less moving parts** — No Docker, no Node.js runtime, no MongoDB
👉 **Agent-first CLI** — JSON output, semantic exit codes, deterministic behavior
👉 **Hybrid storage** — SQLite metadata + file system content
👉 **Two templating options** — Simple includes OR Go's html/template
👉 **Version control** — File system based with instant rollback

## Quick Start

```bash
# Build
go build -o sl-cli ./cmd/sl-cli

# Create site
./sl-cli site create --name "My Site" --slug "my-site"
./sl-cli site version create my-site --version "v1"

# Add content (includes)
./sl-cli site write my-site v1 "index.html" --content '{{>include "nav.html"}}<h1>Home</h1>{{>include "footer.html"}}'

# Or use Go templates
./sl-cli site write my-site v1 "index.html" --content '<h1>{{.title}}</h1>{{range .posts}}<h2>{{.title}}</h2>{{end}}'

# Start daemon
sudo ./sl-cli backend start --daemon --port 3099
curl http://localhost:3099/my-site/
```

## CLI Usage

```bash
# Site management
sl-cli site create --name "Blog" --slug "blog"
sl-cli site version create blog --version "v1"
sl-cli site version switch blog v2

# File management
sl-cli site write blog v1 "index.html" --content "<html>...</html>"
sl-cli site write blog v1 "nav.html" --content "<nav>...</nav>"

# Daemon management
sl-cli backend start --daemon --port 3099
sl-cli backend stop
sl-cli backend stop --uninstall
```

## Architecture

### Core Design
- **Single binary** — No Docker, no runtime dependencies
- **SQLite database** — Embedded, file-based, zero-config
- **Hybrid storage** — SQLite for metadata, file system for content
- **Dynamic blocks** — `{{>include "path"}}` syntax
- **Go templates** — Native html/template with variables, conditionals, loops
- **File system versioning** — Sites stored as `sites/{slug}/{version}/`
- **Agent-first CLI** — JSON output, semantic exit codes
- **Systemd integration** — Auto-installation for boot persistence

### Storage
```
~/.superlandings/
├── db.sql                    # SQLite (metadata)
├── sites/                    # File system (content)
│   └── my-site/
│       ├── v1/
│       │   ├── index.html
│       │   ├── index.html.data.json  # Template data
│       │   └── nav.html
│       └── v2/
└── landings/                 # Legacy support
```

### Dynamic Blocks
```html
{{>include "nav.html"}}
<h1>Welcome</h1>
{{>include "footer.html"}}
```
Recursive includes, processed at serve time, no build step.

### Go Templates
```html
<h1>{{.title}}</h1>
{{if .showBanner}}
  <div>{{.bannerText}}</div>
{{end}}
{{range .posts}}
  <h2>{{.title}}</h2>
{{end}}
```
Data file: `{"title":"My Site","showBanner":true,"bannerText":"Welcome!","posts":[{"title":"Post 1"}]}`

Both approaches can be used together.

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Runtime | Go 1.25+ |
| Database | SQLite (modernc.org/sqlite - pure Go) |
| HTTP | net/http (standard library) |
| CLI | Cobra |
| Storage | Hybrid: SQLite (metadata) + File system (content) |

## Build & Install

```bash
go build -o sl-cli ./cmd/sl-cli
go test ./...

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o sl-cli-linux ./cmd/sl-cli
GOOS=darwin GOARCH=amd64 go build -o sl-cli-macos ./cmd/sl-cli
GOOS=windows GOARCH=amd64 go build -o sl-cli.exe ./cmd/sl-cli
```

## Status

| Capability | State |
|------------|-------|
| Site CRUD operations | ✅ done |
| Site version management | ✅ done |
| Dynamic blocks | ✅ done |
| Go templates | ✅ done |
| File system versioning | ✅ done |
| Hybrid storage | ✅ done |
| Sub-path routing | ✅ done |
| HTTP server | ✅ done |
| Daemon management | ✅ done |
| Systemd auto-installation | ✅ done |
| Agent-first CLI | ✅ done |
| Legacy landing support | ✅ done |
| Traefik integration | ❌ TODO |
| Cloudflare integration | ❌ TODO |
| Blog module | ❌ TODO |
| Organization/user management | ❌ TODO |

## Comparison to Node.js Version

| Version | Runtime | Typical RSS | Startup Time |
|---------|---------|-------------|--------------|
| **SuperLandings Go** | Go (compiled) | ~10 MB | ~100ms |
| SuperLandings (Node.js) | Node.js | ~100-200 MB | ~2-5s |

**Why Go is lighter:**
- No runtime VM — compiled to native machine code
- Single binary — no node_modules, no virtualenv
- Embedded SQLite — no database server process
- Minimal dependencies — only Go stdlib + Cobra

## Links

- **Repository**: https://github.com/javimosch/superlandings-go
- **Original SuperLandings**: https://github.com/javimosch/superlandings
- **Issue Tracker**: https://github.com/javimosch/superlandings-go/issues
- **Vision**: https://github.com/javimosch/superlandings-go/blob/master/docs/VISION.md
- **Roadmap**: https://github.com/javimosch/superlandings-go/blob/master/docs/ROADMAP.md

## License

MIT — <a href="https://www.linkedin.com/in/arancibiajav/" target="_blank">Javier Leandro Arancibia</a>