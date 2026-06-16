# SuperLandings Go - Agent Guide

This document helps AI agents work with the SuperLandings Go codebase.

## Project Overview

Go port of SuperLandings with:
- Single binary deployment
- SQLite (metadata) + file system (content)
- Agent-first CLI with JSON output
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

**Rules:**
1. **NEVER exceed limits** - split immediately if over
2. **Check LOC before committing** - use `wc -l filename`
3. **Split by logical sections** - don't truncate

**Check commands:**
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
# Site management
sl-cli site create --name "Site" --slug "site"
sl-cli site list

# Version management
sl-cli site version create site --version "v1"
sl-cli site version switch site v2

# File management
sl-cli site write site v1 "index.html" --content "<html>...</html>"

# Daemon
sl-cli backend start --daemon --port 3099
sl-cli backend start --daemon --port 3099 --no-systemd
sl-cli backend stop
sl-cli backend stop --uninstall
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

Both work together.

## Architecture Decisions

### Database
- SQLite (modernc.org/sqlite) - pure Go, no CGo
- File: `~/.superlandings/db.sql`
- Suitable for single-server deployments

### Storage
- Hybrid: SQLite (metadata) + file system (content)
- Path: `~/.superlandings/sites/{slug}/{version}/`

### Version Control
- File system based, instant rollback via database flag
- Git-friendly, no migrations

### Templates
- Includes: `{{>include "path"}}` (custom)
- Go templates: html/template with variables/conditionals/loops
- Data files: `.html.data.json` (JSON)
- Process: includes first, then Go template if data file exists

## Agent Memory (Local Skills) 🧠

**IMPORTANT:** Add/update local skills under `~/.agents/skills/` from time to time.

### When to Create Skills
- After learning patterns/caveats
- After implementing features
- After debugging issues
- After identifying recurring problems

### Skill Format
Max 300 LOC, focused:
```
~/.agents/skills/
├── superlandings-go-build/SKILL.md
├── superlandings-go-daemon/SKILL.md
└── superlandings-go-templates/SKILL.md
```

### Skill Content
- Purpose/When to use
- Key commands
- Common patterns
- Caveats/gotchas
- File locations

### Example
```markdown
# SuperLandings Go Build Skill

## Purpose
Build and test the CLI.

## Commands
go build -o sl-cli ./cmd/sl-cli
go test ./...

## Caveats
- Must use Go 1.25+
- Check go.mod if build fails
- Test migrations after schema changes
```

## Common Workflows

### Adding CLI Command
1. Create file in `internal/cli/` (max 400 LOC)
2. Add to root command in `internal/cli/root.go`
3. Implement logic in `internal/services/` (max 400 LOC)
4. Update DB schema in `internal/db/` if needed (max 400 LOC)
5. Test with build and manual invocation
6. Update AGENTS.md if reusable

### Adding Service
1. Create file in `internal/services/` (max 400 LOC)
2. Add repository in `internal/db/repository.go` if needed
3. Add models in `internal/db/models.go` if needed
4. Wire up in CLI
5. Test thoroughly
6. Update docs

### DB Schema Changes
1. Update models in `internal/db/models.go` (max 400 LOC)
2. Add migration to `internal/db/sqlite.go` (max 400 LOC)
3. Test by deleting `~/.superlandings/db.sql` and restarting
4. Update repository if needed
5. Document in BRAINSTORM.md

## Testing

```bash
# Build
go build -o sl-cli ./cmd/sl-cli

# Test
go test ./...

# Manual test
./sl-cli site create --name "Test" --slug "test"
./sl-cli site version create test --version "v1"
./sl-cli site write test v1 "index.html" --content "<h1>Test</h1>"
./sl-cli backend start --daemon --port 3099
curl http://localhost:3099/test
./sl-cli backend stop
```

## Gotchas

### File Size Limits
- **ALWAYS check LOC** before committing
- Split immediately if over limit
- Use `wc -l` to verify

### Database
- File: `~/.superlandings/db.sql`
- Delete to reset: `rm ~/.superlandings/db.sql`
- Migrations auto-run on startup

### Daemon
- PID: `~/.superlandings/sl-cli.pid`
- Log: `~/.superlandings/sl-cli.log`
- Systemd: `/etc/systemd/system/sl-cli.service`
- Use `--no-systemd` to skip auto-install

### Templates
- Includes: `{{>include "path"}}`
- Go templates: `{{.variable}}`, `{{if}}`, `{{range}}`
- Data files: `index.html.data.json` (JSON)
- Process: includes first, then Go template if data file exists

## Contributing

1. Follow file size limits strictly
2. Keep code simple and focused
3. Prefer Go stdlib over external deps
4. Test thoroughly before committing
5. Update docs for new features
6. Create local skills after learning patterns
7. Check LOC before every commit

---

**File size limits are strict. Split files before committing. Local skills are your memory - update them after learnings.**