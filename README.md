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
sl-cli site version create my-site --version "v1" --comment "Initial version"

# Option 1: Simple includes (no data file needed)
sl-cli site write my-site v1 "index.html" --content '{{>include "nav.html"}}<h1>Home</h1>{{>include "footer.html"}}'
sl-cli site write my-site v1 "nav.html" --content '<nav><a href="/my-site/">Home</a></nav>'

# Option 2: Go templates with variables (create .data.json file)
sl-cli site write my-site v1 "index.html" --content '<h1>{{.title}}</h1>{{if .showBanner}}<div>{{.bannerText}}</div>{{end}}{{range .posts}}<h2>{{.title}}</h2>{{end}}'
# Then create index.html.data.json with {"title":"My Site","showBanner":true,"bannerText":"Welcome!","posts":[{"title":"Post 1"}]}

# Start daemon (auto-installs systemd service)
sl-cli backend start --daemon --port 3099

# Access
curl http://localhost:3099/my-site/        # Serves rendered template
```

👉 **Less moving parts** — No Docker, no Node.js runtime, no MongoDB
👉 **Agent-first CLI** — JSON output, semantic exit codes, deterministic behavior
👉 **Hybrid storage** — SQLite for metadata, file system for content
👉 **Two templating options** — Simple includes OR Go's html/template
👉 **Version control** — File system based with instant rollback

## The Problem

The original SuperLandings (Node.js) has many moving parts:
- **Node.js runtime** — Heavy memory footprint, slow startup
- **MongoDB** — Requires database server, complex setup
- **Docker** — Container orchestration overhead
- **npm dependencies** — Large node_modules, security surface area
- **Complex architecture** — Multiple services, async coordination

Without SuperLandings Go, users struggle to:
1. Deploy landing pages without Docker
2. Get fast startup and low memory usage
3. Use agent-friendly CLI interfaces
4. Manage versions without complex database migrations
5. Compose pages from reusable HTML blocks

## The Solution

SuperLandings Go simplifies the stack:
- **Single Go binary** — No runtime, no dependencies, no Docker
- **SQLite database** — Embedded, zero-config, file-based
- **Hybrid storage** — SQLite metadata + file system content
- **Dynamic blocks** — Simple include syntax for template composition
- **Version control** — File system based with instant rollback
- **Agent-first CLI** — JSON output, semantic exit codes, systemd auto-installation

With SuperLandings Go:
1. **Deploy** a single binary — no Docker, no runtime dependencies
2. **Start** instantly — Go compiled binary, fast startup, low memory
3. **Manage** versions via file system — instant rollback, no migrations
4. **Compose** pages with dynamic blocks — `{{>include "nav.html"}}`
5. **Use** agent-friendly CLI — JSON mode, structured errors, deterministic behavior

---

## ⚡ Quick Start

```bash
# Build
go build -o sl-cli ./cmd/sl-cli

# Create a site
./sl-cli site create --name "My Site" --slug "my-site"

# Create version
./sl-cli site version create my-site --version "v1" --comment "Initial version"

# Option 1: Simple includes (no data file needed)
./sl-cli site write my-site v1 "index.html" --content '{{>include "nav.html"}}<h1>Home</h1>{{>include "footer.html"}}'
./sl-cli site write my-site v1 "nav.html" --content '<nav><a href="/my-site/">Home</a> <a href="/my-site/about">About</a></nav>'
./sl-cli site write my-site v1 "footer.html" --content '<footer>&copy; 2025 My Site</footer>'

# Option 2: Go templates with variables (create .data.json file)
./sl-cli site write my-site v1 "index.html" --content '<h1>{{.title}}</h1>{{if .showBanner}}<div>{{.bannerText}}</div>{{end}}{{range .posts}}<h2>{{.title}}</h2>{{end}}'
# Create index.html.data.json: {"title":"My Site","showBanner":true,"bannerText":"Welcome!","posts":[{"title":"Post 1"}]}

# Start daemon (auto systemd installation)
sudo ./sl-cli backend start --daemon --port 3099

# Access
curl http://localhost:3099/my-site/
curl http://localhost:3099/my-site/about

# Stop daemon
./sl-cli backend stop

# Stop and uninstall systemd service
./sl-cli backend stop --uninstall
```

---

## For Humans

| Instead of... | You do... |
|--------------|-----------|
| Docker containers | Single Go binary |
| MongoDB setup | SQLite embedded in binary |
| npm install | `go build` — no dependencies |
| Complex deployment | Copy binary to server |
| Database migrations | File system versioning |

What this means day-to-day:
- **No Docker** — Just copy the binary and run it
- **No MongoDB** — SQLite file is created automatically
- **No npm** — Go standard library only
- **No dependencies** — Single binary, no external deps
- **Fast startup** — Go compiled binary starts instantly

## For AI Agents

- 🔍 **Deterministic** — JSON output with `--json`, semantic exit codes
- 🛠️ **Direct service calls** — CLI commands hit business logic directly (no HTTP API)
- 💾 **Hybrid storage** — SQLite metadata + file system content
- 🎯 **Version control** — File system based with instant rollback
- 🚀 **Systemd auto-install** — `--daemon` flag auto-installs systemd service
- 📦 **Single binary** — No runtime dependencies, easy deployment
- 🔄 **Dynamic blocks** — `{{>include "path"}}` syntax for template composition

```bash
# Agent workflow: create site -> version -> files -> daemon -> access
sl-cli site create --name "X" --slug "x"        # JSON response
sl-cli site version create x --version "v1"       # JSON response
sl-cli site write x v1 "index.html" --content "..." # JSON response
sl-cli backend start --daemon --port 3099       # Auto systemd install
curl http://localhost:3099/x/                    # Access site
```

---

## What You Get

SuperLandings Go gives you a simplified landing page management system:

- 🎯 **Single binary** — No Docker, no Node.js runtime, no dependencies
- 📦 **SQLite database** — Embedded, zero-config, file-based persistence
- 🔄 **Dynamic blocks** — Simple include syntax for template composition
- 📁 **File system versioning** — Instant rollback, no migrations
- 🚀 **Systemd auto-install** — Boot-persistent daemon with single command
- 🔍 **Agent-first CLI** — JSON output, semantic exit codes, deterministic behavior
- 🌐 **Sub-path routing** — `/site/page` serves `page.html` with dynamic blocks
- 💾 **Hybrid storage** — SQLite metadata + file system content (best of both worlds)

---

## 🛠️ CLI Usage Examples

```bash
# Site management
sl-cli site list                                    # List all sites
sl-cli site create --name "Blog" --slug "blog"  # Create site
sl-cli site version create blog --version "v1"     # Create version
sl-cli site version list blog                       # List versions
sl-cli site version switch blog v2                 # Switch active version

# File management
sl-cli site write blog v1 "index.html" --content "<html>...</html>"
sl-cli site write blog v1 "nav.html" --content "<nav>...</nav>"
sl-cli site write blog v1 "footer.html" --content "<footer>...</footer>"

# Daemon management
sl-cli backend start --daemon --port 3099           # Start with systemd auto-install
sl-cli backend status                               # Check daemon status
sl-cli backend stop                                # Stop daemon
sl-cli backend stop --uninstall                     # Stop + remove systemd service

# Legacy landing support
sl-cli landing create --name "Landing" --slug "landing" --type html --content "<html>...</html>"
sl-cli landing list                                 # List landings
sl-cli backend start --daemon --port 3099           # Serves both sites and landings
```

---

## 🏗️ Architecture

### Core Design

SuperLandings Go is a **simplified Go port** of the Node.js SuperLandings:

- **Single binary** — No Docker, no runtime dependencies
- **SQLite database** — Embedded, file-based, zero-config
- **Hybrid storage** — SQLite for metadata, file system for content
- **Dynamic blocks** — `{{>include "path"}}` syntax processed at serve time
- **File system versioning** — Sites stored as `sites/{slug}/{version}/`
- **Agent-first CLI** — JSON output, semantic exit codes, deterministic behavior
- **Systemd integration** — Auto-installation for boot persistence

### Storage Architecture

```
~/.superlandings/
├── db.sql                    # SQLite database (metadata)
├── sites/                    # File system (content)
│   └── my-site/
│       ├── v1/
│       │   ├── index.html
│       │   ├── nav.html
│       │   └── footer.html
│       └── v2/
│           ├── index.html
│           ├── nav.html
│           └── footer.html
└── landings/                 # Legacy landing support
    └── test-landing/
        └── index.html
```

**SQLite stores:**
- Sites (id, name, slug, created_at, updated_at)
- Site versions (id, site_id, version, path, comment, author, is_active, created_at)
- Landings (legacy support)

**File system stores:**
- Actual HTML files for each version
- Dynamic block files (nav.html, footer.html, etc.)

### Dynamic Blocks

Include syntax processed at serve time:

```html
<!-- index.html -->
{{>include "nav.html"}}
<h1>Welcome</h1>
{{>include "footer.html"}}
```

**Features:**
- Recursive includes (files can include other files)
- Relative to version directory
- Processed on each request (no build step)
- Works with sub-path routing (`/site/page`)

### Go Templates

Native Go `html/template` support with variables, conditionals, and loops:

```html
<!-- index.html -->
<h1>{{.title}}</h1>
{{if .showBanner}}
  <div class="banner">{{.bannerText}}</div>
{{end}}
{{range .posts}}
  <article>
    <h2>{{.title}}</h2>
    <p>{{.content}}</p>
  </article>
{{end}}
```

**Data file (index.html.data.json):**
```json
{
  "title": "My Blog",
  "showBanner": true,
  "bannerText": "Welcome!",
  "posts": [
    {"title": "Post 1", "content": "..."},
    {"title": "Post 2", "content": "..."}
  ]
}
```

**Features:**
- Native Go template engine (built-in, secure)
- Variables: `{{.variable}}`
- Conditionals: `{{if .condition}}...{{end}}`
- Loops: `{{range .items}}...{{end}}`
- Auto-escapes HTML (XSS protection)
- Optional — works without data file (just includes)

**Two Approaches:**
1. **Simple includes** — Use `{{>include "path"}}` for basic composition
2. **Go templates** — Create `.data.json` file for variables, conditionals, loops

Both can be used together: Go templates can include files, and includes can contain Go template syntax.

### Version Control

File system based versioning:

```bash
# Create v1
sl-cli site version create my-site --version "v1"
sl-cli site write my-site v1 "index.html" --content "<h1>v1</h1>"

# Create v2
sl-cli site version create my-site --version "v2"
sl-cli site write my-site v2 "index.html" --content "<h1>v2</h1>"

# Switch to v2
sl-cli site version switch my-site v2

# Rollback to v1
sl-cli site version switch my-site v1
```

**Benefits:**
- Instant rollback (symlink or database flag)
- Easy backup (tar, rsync)
- Git-friendly
- No database migrations

### Systemd Integration

Automatic systemd service installation:

```bash
# Start daemon (auto-installs systemd service)
sudo sl-cli backend start --daemon --port 3099

# This automatically:
# 1. Creates /etc/systemd/system/sl-cli.service
# 2. Runs systemctl daemon-reload
# 3. Enables service (auto-start on boot)
# 4. Starts service
# 5. Logs to journald

# Stop and uninstall
sl-cli backend stop --uninstall
```

**Features:**
- Auto-detects systemd availability
- Falls back to basic daemon if systemd unavailable
- `--no-systemd` flag to disable auto-installation
- Log management via journald
- Auto-restart on crashes

---

## 📤 Output Envelope + Exit Codes

### Output Modes

| Mode | Command | Output |
|------|---------|--------|
| **text** (default) | `sl-cli site list` | Table format |
| **json** | `sl-cli site list --json` | JSON array |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Success |
| `80-89` | User errors (invalid input, permission denied) |
| `90-99` | Resource errors (not found, already exists) |
| `100-109` | Integration errors (network, timeout) |
| `110-119` | Software errors (internal, unexpected) |

---

## ⚙️ Configuration

### Config File

No config file required — all configuration is CLI flags or environment variables.

### Environment Variables

| Variable | Purpose | Default |
|----------|---------|---------|
| `SUPERLANDINGS_UI_DIR` | UI override directory for development | — |

### File Paths

| Path | Purpose |
|------|---------|
| `~/.superlandings/db.sql` | SQLite database |
| `~/.superlandings/sites/` | Site file system storage |
| `~/.superlandings/landings/` | Legacy landing file system storage |
| `~/.superlandings/sl-cli.pid` | Daemon PID file |
| `~/.superlandings/sl-cli.log` | Daemon log file |

---

## 📦 Build & Install

```bash
# Build
go build -o sl-cli ./cmd/sl-cli

# Run tests
go test ./...

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o sl-cli-linux ./cmd/sl-cli

# Cross-compile for macOS
GOOS=darwin GOARCH=amd64 go build -o sl-cli-macos ./cmd/sl-cli

# Cross-compile for Windows
GOOS=windows GOARCH=amd64 go build -o sl-cli.exe ./cmd/sl-cli
```

Requires Go 1.25+.

---

## 🔧 Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| Port already in use | Another process using port 3099 | Use `--port 3000` |
| Permission denied | No write access to `~/.superlandings/` | Check directory permissions |
| systemd not found | System not using systemd | Use `--no-systemd` flag |
| Include not found | File path in `{{>include}}` doesn't exist | Check file path and version |

---

## 🧱 Tech Stack

| Layer | Technology |
|-------|-----------|
| Runtime | Go 1.25+ |
| Database | SQLite (modernc.org/sqlite - pure Go, no CGo) |
| HTTP | net/http (standard library) |
| CLI | Cobra |
| Storage | Hybrid: SQLite (metadata) + File system (content) |
| Output | Text (default) / JSON (with `--json`) |
| Process management | systemd (auto-install) or basic daemon |

---

## ⚡ Performance

SuperLandings Go is designed to be lightweight — Go's compiled nature and minimal dependencies result in fast startup and low memory usage compared to the Node.js version.

### Resource Usage (estimated)

| Operation | Max RSS | Startup Time |
|-----------|---------|--------------|
| CLI help | ~5 MB | <10ms |
| Site creation | ~8 MB | ~50ms |
| Version creation | ~8 MB | ~50ms |
| Daemon startup | ~10 MB | ~100ms |
| HTTP serving | ~12 MB | N/A |

### Comparison to Node.js Version

| Version | Runtime | Typical RSS | Startup Time |
|---------|---------|-------------|--------------|
| **SuperLandings Go** | Go (compiled) | ~10 MB | ~100ms |
| SuperLandings (Node.js) | Node.js | ~100-200 MB | ~2-5s |

**Why SuperLandings Go is lighter:**
- **No runtime VM** — Go compiles to native machine code
- **Single binary** — No node_modules, no virtualenv
- **Embedded SQLite** — No database server process
- **Minimal dependencies** — Only Go standard library + Cobra

---

## 🌐 Status

| Capability | State |
|------------|-------|
| Site CRUD operations | ✅ done |
| Site version management (create/list/switch) | ✅ done |
| Dynamic blocks (`{{>include "path"}}`) | ✅ done |
| Go templates (html/template) | ✅ done |
| File system versioning | ✅ done |
| Hybrid storage (SQLite + FS) | ✅ done |
| Sub-path routing (`/site/page`) | ✅ done |
| HTTP server with serving | ✅ done |
| Daemon management | ✅ done |
| Systemd auto-installation | ✅ done |
| Agent-first CLI (JSON output, exit codes) | ✅ done |
| Legacy landing support | ✅ done |
| Traefik integration | ❌ TODO |
| Cloudflare integration | ❌ TODO |
| Blog module | ❌ TODO |
| Organization/user management | ❌ TODO |

---

## 📝 Changelog

- [v1.0.0 — Initial Go port with sites, versioning, and dynamic blocks](#)

---

## 🔄 Comparison with Original SuperLandings

| Feature | Node.js Version | Go Version |
|---------|----------------|-----------|
| Runtime | Node.js | Go (compiled) |
| Database | MongoDB + JSON | SQLite |
| Deployment | Docker | Single binary |
| Dependencies | npm (hundreds) | Go stdlib + Cobra |
| Startup time | 2-5s | <100ms |
| Memory usage | 100-200 MB | ~10 MB |
| Version control | Database | File system |
| Dynamic blocks | ❌ No | ✅ Yes |
| Agent-first CLI | ❌ No | ✅ Yes |
| Systemd integration | Manual | Auto-installation |

**Key improvements in Go version:**
- ✅ **Simpler deployment** — No Docker, no database server
- ✅ **Faster startup** — Compiled binary vs interpreted runtime
- ✅ **Lower memory** ~10-20x less memory usage
- ✅ **Dynamic blocks** — Template composition without build step
- ✅ **File system versioning** — Instant rollback, no migrations
- ✅ **Agent-first CLI** — JSON output, semantic exit codes
- ✅ **Systemd auto-installation** — One-command boot persistence

---

## 🔗 Links

- **Repository**: https://github.com/javimosch/superlandings-go
- **Original SuperLandings**: https://github.com/javimosch/superlandings
- **Issue Tracker**: https://github.com/javimosch/superlandings-go/issues

---

## License

MIT — <a href="https://www.linkedin.com/in/arancibiajav/" target="_blank">Javier Leandro Arancibia</a>