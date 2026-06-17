# SuperLandings Go - Brainstorming & Design Notes

## Project Overview

Port SuperLandings (Node.js) to Go with simplified architecture, focusing on:
- Single binary deployment
- SQLite instead of MongoDB
- File system-based versioning
- Agent-first CLI
- Minimal dependencies

## Architecture Decisions

### Database
- **Choice**: SQLite (modernc.org/sqlite - pure Go, no CGo)
- **Reason**: Embedded, zero-config, file-based, suitable for single-server deployments
- **Trade-off**: Not suitable for high-concurrency multi-tenant, but perfect for target use case

### Storage Strategy
- **Hybrid approach**: SQLite for metadata, file system for content
- **SQLite stores**: Sites, versions, users, organizations, domains
- **File system stores**: Actual HTML files, assets, templates
- **Benefits**: Best of both worlds - queries via SQL, content via FS

### Version Control
- **File system based**: `sites/{slug}/{version}/`
- **Instant rollback**: Database flag or symlink
- **Git-friendly**: Can version control entire sites
- **No migrations**: File system doesn't need schema changes

### Dynamic Blocks
- **Syntax**: `{{>include "path"}}`
- **Processed at**: Serve time (no build step)
- **Recursive**: Files can include other files
- **Alternative**: Go's html/template for variables, conditionals, loops

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

### Landing (legacy support)
```go
type Landing struct {
    ID             string
    Name           string
    Slug           string
    Type           string    // html, virtual, static
    OrganizationID string
    Content        string
    Files          []File
    Domains        []Domain
    Config         Config
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

## CLI Structure

### Root Commands
- `sl-cli landing` - Landing CRUD
- `sl-cli backend` - Daemon management
- `sl-cli site` - Site management
- `sl-cli organization` - Organization CRUD
- `sl-cli user` - User management

### Site Commands
- `sl-cli site create --name --slug`
- `sl-cli site list`
- `sl-cli site version create <site> --version --comment`
- `sl-cli site version list <site>`
- `sl-cli site version switch <site> <version>`
- `sl-cli site write <site> <version> <file> --content`

### Backend Commands
- `sl-cli backend start --daemon --port [--no-systemd]`
- `sl-cli backend stop [--uninstall]`
- `sl-cli backend status`

## Daemon Implementation

### Features
- PID file management
- Process lifecycle (start, stop, status)
- Log redirection to file
- Systemd auto-installation
- Fallback to basic daemon if systemd unavailable

### Systemd Service
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
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

## HTTP Server

### Routes
- `/:slug` - Serve site index
- `/:slug/:page` - Serve site page with sub-path routing
- `/` - List all sites and landings
- `/health` - Health check

### Serving Logic
1. Try to serve as site (with dynamic blocks and Go templates)
2. Fall back to landing
3. Return 404 if not found

### Template Processing
1. Process includes (`{{>include "path"}}`)
2. If `.data.json` exists, render with Go html/template
3. Return processed HTML

## File System Structure

```
~/.superlandings/
├── db.sql                    # SQLite database
├── sites/                    # Site content
│   └── my-site/
│       ├── v1/
│       │   ├── index.html
│       │   ├── index.html.data.json  # Template variables
│       │   ├── nav.html
│       │   └── footer.html
│       └── v2/
├── landings/                 # Legacy landing content
│   └── test-landing/
│       └── index.html
├── sl-cli.pid               # Daemon PID file
└── sl-cli.log               # Daemon log file
```

## Implementation Phases

### Phase 1: Core ✅
- SQLite database layer
- Site CRUD operations
- Site version management
- File system operations
- HTTP server
- Daemon management
- Dynamic blocks
- Go templates

### Phase 2: Core Infrastructure (Next)
- Traefik integration
- Cloudflare integration
- Domain management

### Phase 3: Content Management
- Blog module
- Asset management
- SEO meta tags

### Phase 4: Multi-Tenancy
- User management
- Organization management
- Authentication

## Future Considerations

### Potential Features
- Admin UI (React embedded via go:embed)
- Database backups
- Multi-region deployment
- GraphQL API
- WebAssembly templates

### Architecture Decisions to Revisit
- SQLite vs PostgreSQL for high-traffic multi-tenant
- Single binary vs microservices for scaling
- File system vs S3 for cloud-native storage

## Notes

- Keep files under 500 LOC (per global rules)
- Prefer Go stdlib over external dependencies
- Agent-first design: JSON output, semantic exit codes, deterministic behavior
- Simple over complex: 20% of features covering 80% of use cases