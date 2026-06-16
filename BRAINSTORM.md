# SuperLandings Go Port - Architecture Brainstorm

## Overview
Port SuperLandings from Node.js to Go, focusing on `sl-cli` (CLI+UI) with direct backend logic access, SQLite persistence, and single-binary deployment.

## Key Requirements
1. **sl-cli (Go version)** - CLI commands with embedded React UI
2. **sl-cli backend start/stop --daemon** - Process management
3. **Direct backend logic** - No HTTP API for CLI commands, hit logic directly
4. **UI override** - Load .html from DIR for faster iterations when deployed
5. **SQLite DB** - Replace MongoDB with `~/.superlandings/db.sql`
6. **Single binary** - No Docker, deploy as standalone binary

## Current Node.js Architecture Analysis

### CLI Commands (`cli/lib/commands/`)
- **landing.js** (704 LOC) - Landing CRUD operations
  - list, get, create, update, delete
  - content update, domain management
  - version management
- **organization.js** - Organization management
- **user.js** - User management

### Persistence Layer (`lib/store.js`)
- **Dual engine**: JSON file + MongoDB
- **Collections**: landings, versions, app_state
- **Sync logic**: JSON ↔ MongoDB bidirectional sync

### Core Libraries (`lib/`)
- **blog.js** (14k LOC) - Blog module with posts
- **traefik.js** (17k LOC) - Traefik dynamic configuration
- **cloudflare.js** (11k LOC) - Cloudflare DNS integration
- **versions.js** (14k LOC) - Version control system
- **llm.js** (19k LOC) - AI/LLM integration
- **audit.js** - Audit logging
- **auth.js** - Authentication
- **i18n.js** - Internationalization

## Proposed Go Architecture

### Project Structure
```
superlandings-go/
├── cmd/
│   └── sl-cli/
│       └── main.go           # CLI entry point
├── internal/
│   ├── cli/                  # CLI command implementations
│   │   ├── landing.go        # Landing commands
│   │   ├── organization.go   # Organization commands
│   │   ├── user.go           # User commands
│   │   └── backend.go        # Backend daemon commands
│   ├── db/                   # SQLite database layer
│   │   ├── sqlite.go         # SQLite connection & migrations
│   │   ├── models.go         # Data models
│   │   └── repositories.go   # Database operations
│   ├── services/             # Business logic (direct access)
│   │   ├── landing.go        # Landing service
│   │   ├── blog.go           # Blog service
│   │   ├── traefik.go        # Traefik service
│   │   ├── cloudflare.go     # Cloudflare service
│   │   └── version.go        # Version service
│   ├── daemon/               # Process management
│   │   └── daemon.go         # PID file, signals, lifecycle
│   ├── server/               # HTTP server for UI
│   │   ├── server.go         # HTTP handlers
│   │   └── embed.go          # Embedded filesystem
│   └── config/               # Configuration
│       └── config.go         # Config loading
├── ui/                       # Frontend (overrideable)
│   ├── index.html            # React 18 entry
│   ├── css/
│   │   └── app.css
│   └── js/
│       ├── app.jsx
│       ├── components/
│       └── views/
├── go.mod
├── go.sum
├── build.sh
└── README.md
```

## Phase 1: Core Infrastructure (MVP)

### 1.1 SQLite Database Layer
**File**: `internal/db/sqlite.go`
- Use `modernc.org/sqlite` (pure Go, no CGo)
- Database path: `~/.superlandings/db.sql`
- Auto-migration on startup
- Connection pooling

**Models** (`internal/db/models.go`):
```go
type Landing struct {
    ID             string    `json:"id" db:"id"`
    Name           string    `json:"name" db:"name"`
    Slug           string    `json:"slug" db:"slug"`
    Type           string    `json:"type" db:"type"` // html, ejs, virtual, static
    OrganizationID string    `json:"organizationId" db:"organization_id"`
    Content        string    `json:"content" db:"content"` // for html type
    Files          []File    `json:"files" db:"-"` // for virtual type
    Domains        []Domain  `json:"domains" db:"-"` 
    Config         Config    `json:"config" db:"config"`
    CreatedAt      time.Time `json:"createdAt" db:"created_at"`
    UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

type Domain struct {
    Domain     string `json:"domain" db:"domain"`
    Traefik    bool   `json:"traefik" db:"traefik"`
    Cloudflare bool   `json:"cloudflare" db:"cloudflare"`
}

type File struct {
    Path string `json:"path" db:"path"`
    Content string `json:"content" db:"content"`
}
```

### 1.2 CLI Framework
**File**: `cmd/sl-cli/main.go`
- Use `github.com/spf13/cobra` for CLI commands
- JSON output mode for agent-friendly usage
- Semantic exit codes (matching Node.js version)

**Command Structure**:
```bash
sl-cli landing list          # Direct DB call
sl-cli landing get <id>      # Direct DB call
sl-cli landing create ...    # Direct service call
sl-cli backend start         # Start HTTP server
sl-cli backend stop          # Stop daemon
sl-cli backend status        # Check daemon status
```

### 1.3 Business Logic Layer
**Key Design**: CLI commands call services directly (no HTTP)

```go
// internal/services/landing.go
type LandingService struct {
    db *sqlite.DB
}

func (s *LandingService) CreateLanding(req CreateLandingRequest) (*Landing, error) {
    // Direct database operations
    // File system operations
    // Validation
}

// internal/cli/landing.go
func (c *LandingCommand) Create(cmd *cobra.Command, args []string) {
    service := services.NewLandingService(c.db)
    landing, err := service.CreateLanding(req)
    // Output JSON or error
}
```

### 1.4 Daemon Management
**File**: `internal/daemon/daemon.go`
- PID file: `~/.superlandings/sl-cli.pid`
- Log file: `~/.superlandings/sl-cli.log`
- Signal handling (SIGTERM, SIGINT)
- Status checking

```go
func StartDaemon(port int) error
func StopDaemon() error
func DaemonStatus() (bool, int, error)
```

### 1.5 HTTP Server for UI
**File**: `internal/server/server.go`
- Embedded UI via `go:embed`
- Override mode: Serve from directory if flag provided
- JSON API endpoints for UI
- Static file serving

**Override Mechanism**:
```go
func startServer(uiDir string) {
    var fileServer http.Handler
    
    if uiDir != "" {
        // Serve from disk for development
        fileServer = http.FileServer(http.Dir(uiDir))
    } else {
        // Serve embedded files
        uiSub, _ := fs.Sub(uiFiles, "ui")
        fileServer = http.FileServer(http.FS(uiSub))
    }
}
```

## Phase 2: Core Landing Features

### 2.1 Landing CRUD
- **Create**: Support html, ejs, virtual, static types
- **Read**: List, get by ID/slug
- **Update**: Content, metadata, files
- **Delete**: Soft delete with cleanup

### 2.2 File System Operations
- Landing directories: `~/.superlandings/landings/{slug}/`
- Virtual file management
- Asset upload handling
- Directory structure validation

### 2.3 Domain Management
- Add/remove domains
- Traefik configuration generation
- Cloudflare DNS integration (Phase 3)

## Phase 3: Advanced Features

### 3.1 Traefik Integration
**File**: `internal/services/traefik.go`
- Generate dynamic configuration
- Watch for changes
- Hot-reload Traefik
- SSL certificate management

### 3.2 Cloudflare Integration
**File**: `internal/services/cloudflare.go`
- DNS record management
- Proxy configuration
- API authentication

### 3.3 Version Control
**File**: `internal/services/version.go`
- Version creation
- Restore functionality
- Version metadata
- Storage optimization

### 3.4 Blog Module
**File**: `internal/services/blog.go`
- Post CRUD
- SEO metadata
- Multi-language support
- RSS feed generation

## Technical Decisions

### Database: SQLite
**Pros**:
- Single file, no external dependencies
- Perfect for single-binary deployment
- Good performance for CLI use case
- Easy backup/restore

**Libraries**:
- `modernc.org/sqlite` - Pure Go, no CGo
- `gorm.io/gorm` - ORM (optional, can use raw SQL)

### CLI Framework: Cobra
**Pros**:
- Standard for Go CLIs
- Good documentation
- Subcommand structure
- Auto-generated help

### HTTP Server: Standard Library
**Pros**:
- No dependencies
- Sufficient for UI needs
- Good performance

### Embed: Go 1.16+
**Pros**:
- Single binary
- No runtime file dependencies
- Override mode for development

## File Size Management (500 LOC Limit)

### Strategy:
- Split large files into focused modules
- Use interfaces for abstraction
- Separate concerns (CLI, service, repository)
- Keep handlers thin

### Example Splits:
```
landing.go (500 LOC max)
├── landing_crud.go       # CRUD operations
├── landing_files.go      # File operations
├── landing_domains.go    # Domain management
└── landing_validation.go # Validation logic
```

## Development Workflow

### 1. Development Mode
```bash
# Start server with UI override for hot-reload
sl-cli backend start --ui-dir ./ui

# Edit React files in ./ui/
# Refresh browser - no recompile needed
```

### 2. Production Build
```bash
# Build single binary
./build.sh

# Deploy
./sl-cli backend start --daemon
```

### 3. CLI Usage
```bash
# Direct service calls (no HTTP)
sl-cli landing list
sl-cli landing create --name "My Landing" --slug "my-landing" --type html
sl-cli landing update my-landing --content "<html>...</html>"
```

## Migration Strategy

### Data Migration (Node.js → Go)
1. Export MongoDB/JSON to intermediate format
2. Import into SQLite
3. Validate data integrity
4. Migrate file system assets

### Feature Parity Checklist
- [ ] Landing CRUD (html, ejs, virtual, static)
- [ ] Domain management
- [ ] Traefik integration
- [ ] Cloudflare DNS
- [ ] Version control
- [ ] Blog module
- [ ] User/organization management
- [ ] Audit logging
- [ ] Authentication
- [ ] i18n support

## Performance Considerations

### SQLite Optimization
- WAL mode for concurrent access
- Connection pooling
- Prepared statements
- Indexes on frequently queried fields

### Binary Size
- Use `upx` for compression (optional)
- Strip debug symbols in production
- Minimize dependencies

## Testing Strategy

### Unit Tests
- Service layer logic
- Repository operations
- CLI command handlers

### Integration Tests
- Database operations
- File system interactions
- HTTP endpoints

### CLI Tests
- Command execution
- Exit codes
- JSON output validation

## Next Steps

1. **Setup project structure** - Initialize Go module, directories
2. **Implement SQLite layer** - Models, migrations, repositories
3. **Build CLI framework** - Cobra setup, basic commands
4. **Create landing service** - Core CRUD operations
5. **Add daemon management** - Start/stop/status
6. **Implement HTTP server** - Embedded UI, override mode
7. **Port UI from Node.js** - React components, views
8. **Add advanced features** - Traefik, Cloudflare, versions
9. **Testing & optimization** - Unit tests, performance tuning
10. **Documentation** - AGENTS.md, README

## Questions for User

1. **Priority order**: Should we focus on landing CRUD first, or daemon/HTTP server?
2. **Feature scope**: Do you need all features (blog, traefik, cloudflare) in MVP?
3. **Data migration**: Any existing data that needs migration?
4. **Authentication**: Is auth needed for MVP or can it be deferred?
5. **UI fidelity**: Should the Go UI match Node.js exactly, or can we simplify?