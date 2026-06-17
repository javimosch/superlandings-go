<p align="center">
  <img src="https://img.shields.io/badge/version-1.0.0-blue" alt="Version">
  <img src="https://img.shields.io/badge/license-MIT-green" alt="License">
  <img src="https://img.shields.io/badge/go-1.25+-orange" alt="Go">
</p>

<h1 align="center">SuperLandings Go 🚀</h1>

<p align="center">
  <b>Agent-first static site manager with versioning, assets, and domain serving.</b><br>
  <b>Single binary, zero dependencies, JSON-native CLI.</b>
</p>

> Single binary landing page management with versioning, dynamic blocks, Go templates,
> shared assets, domain-aware serving, and SQLite persistence.

```bash
# Create a site
sl-cli site create --name "My Site" --slug "my-site"
sl-cli site version create my-site --version "v1"
sl-cli site write my-site v1 "index.html" --content '<h1>{{.title}}</h1>'
sl-cli site write my-site v1 "nav.html" --content '<nav>{{>include "nav.html"}}</nav>'

# Upload assets (shared across versions)
sl-cli site upload my-site "logo.png" --file ./logo.png
sl-cli site upload my-site "css/style.css" --file ./dist/style.css

# Reference assets in templates
echo '<img src="{{asset "logo.png"}}" alt="Logo">' | sl-cli site write my-site v1 "index.html" --content "$(cat)"

# Start daemon
sl-cli backend start --daemon --port 3099
curl http://localhost:3099/my-site/
```

- 👉 **Agent-first** — JSON by default, semantic exit codes, versioned output
- 👉 **No Docker, no Node, no MongoDB** — compiled Go binary + SQLite
- 👉 **Shared assets** — images, CSS, JS stored once, shared across versions
- 👉 **{{asset "file"}}** — template helper resolves assets by filename
- 👉 **Admin panel** — schema-driven UI, blog editor (EasyMDE), raw HTML (CodeMirror), form editor
- 👉 **Domain-aware** — serves from Host header, including root path
- 👉 **Remote execution** — all commands support `--target <host:port>`

## Quick Start

```bash
go build -o sl-cli ./cmd/sl-cli

# Create and serve a site
./sl-cli site create --name "Blog" --slug "blog"
./sl-cli site version create blog --version "v1"
./sl-cli site write blog v1 "index.html" --content '<h1>Hello</h1>'
./sl-cli backend start --daemon --port 3099
curl http://localhost:3099/blog/
```

## CLI Usage

```bash
# Sites
sl-cli site create --name "Site" --slug "site"
sl-cli site list
sl-cli site version create site --version "v1"
sl-cli site version switch site v2
sl-cli site write site v1 "file.html" --content "<html>..."

# Assets (shared across all versions)
sl-cli site upload site "logo.png" --file ./logo.png
sl-cli site assets list site
sl-cli site assets remove site "path/asset.png"

# Admin
sl-cli site admin create <site>           # Generate token
sl-cli site admin configure <site> --auto-detect  # Generate schema
sl-cli user create --email user --password pass  # Create user
sl-cli user grant <site> <email>          # Grant site access

# Templates
# {{>include "nav.html"}}        — recursive include
# {{.variable}} / {{if}}/{{range}} — Go template
# {{asset "logo.png"}}           — resolve asset by filename

# DNS & Traefik (via hotify-cli on remote)
sl-cli site dns setup site --domain site.example.com --ip 1.2.3.4 --traefik

# Remote execution
sl-cli site list --target dk2
sl-cli site upload site "img.png" --file ./img.png --target dk2

# Daemon
sl-cli backend start --daemon --port 3099
sl-cli backend stop
sl-cli backend status
```

## Admin Panel

The admin panel is **opt-in** and **schema-driven**. Enable it per site:

### Auth Modes
| Mode | URL | Behavior |
|------|-----|----------|
| `none` (default) | `/admin/{slug}/{token}` | Token-gated, no login |
| `password` | `/admin/{slug}` | Login form, JWT sessions, logout |

### Editor Types
| Section | Use Case | Editor |
|---------|----------|--------|
| `form` + `.html` source | Raw HTML files | CodeMirror (syntax highlight, line numbers) |
| `form` + `.data.json` source | Template data fields | Form fields (text, textarea) |
| `markdown` + `blog` dir | Blog posts | EasyMDE (markdown, metadata, drafts) |

### Schema Example
```json
{
  "auth": "password",
  "sections": [
    {"title":"Page","type":"form","source":"index.html.data.json","fields":[
      {"key":"hero_title","label":"Hero Title","type":"text"}
    ]},
    {"title":"Blog","type":"markdown","source":"blog"},
    {"title":"HTML","type":"form","source":"index.html",
     "layout":{"editorHeight":"calc(90vh - 50px)"},"fields":[
      {"key":"_raw_html","label":"Full HTML","type":"textarea"}
    ]}
  ]
}
```

## Architecture

### Storage

```
~/.superlandings/
├── db.sql                    # SQLite (metadata: sites, versions, domains, users)
├── sites/{slug}/
│   ├── assets/               # Shared across versions (images, CSS, JS)
│   │   ├── logo.png
│   │   ├── css/style.css
│   │   └── img/photo.jpg
│   ├── v1/                   # Versioned content
│   │   ├── index.html
│   │   ├── index.html.data.json
│   │   ├── nav.html
│   │   └── pages/about.html
│   └── v2/                   # Rollback-ready
├── sl-cli.pid
└── sl-cli.log
```

### Serving

| Mode | URL | Resolution |
|------|-----|-----------|
| Path-based | `localhost:3099/slug/path` | Extracts slug from URL path |
| Domain-based | `test.domain.com/path` | Looks up slug from site_domains via Host header |

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Runtime | Go 1.25+ |
| Database | SQLite (modernc.org/sqlite - pure Go) |
| HTTP | net/http (standard library) |
| CLI | Cobra |
| DNS/Traefik | hotify-cli (pluggable) |

## Status

| Capability | State |
|------------|-------|
| Site CRUD + versioning | ✅ |
| Dynamic blocks + Go templates | ✅ |
| {{asset "file"}} asset resolver | ✅ |
| Asset upload / list / remove | ✅ |
| Domain-aware serving (Host header) | ✅ |
| Agent-first CLI (JSON, semantic codes) | ✅ |
| Remote execution (--target) | ✅ |
| Hybrid SQLite + FS storage | ✅ |
| Daemon + systemd | ✅ |
| DNS/Traefik (via hotify-cli) | ✅ |
| Legacy landing support | ✅ |
| Blog module | ✅ |
| Multi-tenancy | 🔜 |

## Build & Install

```bash
go build -o sl-cli ./cmd/sl-cli

# Cross-compile
GOOS=linux GOARCH=amd64 go build -o sl-cli-linux ./cmd/sl-cli
GOOS=darwin GOARCH=amd64 go build -o sl-cli-macos ./cmd/sl-cli
```

## Links

- **Repository**: https://github.com/javimosch/superlandings-go
- **Original**: https://github.com/javimosch/superlandings (Node.js)
- **Vision**: ./docs/vision.md
- **Roadmap**: ./docs/roadmap.md

## License

MIT — <a href="https://www.linkedin.com/in/arancibiajav/" target="_blank">Javier Leandro Arancibia</a>
