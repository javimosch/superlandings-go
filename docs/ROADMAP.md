# Roadmap 🗺️

This roadmap outlines the planned development of SuperLandings Go, aligned with our vision of frictionless, agent-first static site creation with dynamic capabilities.

## Current Status (v1.0.0)

### ✅ Completed Features
- **Site Management** - Create, list sites
- **Version Control** - File system based versioning with instant rollback
- **Dynamic Blocks** - `{{>include "path"}}` syntax for template composition
- **Hybrid Storage** - SQLite (metadata) + File system (content)
- **Sub-path Routing** - `/site/page` serves `page.html` with dynamic blocks
- **HTTP Server** - Serve sites and legacy landings
- **Daemon Management** - Background process with PID file
- **Systemd Integration** - Auto-installation for boot persistence
- **Agent-First CLI** - JSON output, semantic exit codes, deterministic behavior
- **Legacy Landing Support** - Basic HTML landing pages

### 🎯 What Works Today
```bash
# Create a site with versioning
sl-cli site create --name "My Site" --slug "my-site"
sl-cli site version create my-site --version "v1"

# Add pages with dynamic blocks
sl-cli site write my-site v1 "index.html" --content '{{>include "nav.html"}}<h1>Home</h1>{{>include "footer.html"}}'

# Deploy instantly
sl-cli backend start --daemon --port 3099
curl http://localhost:3099/my-site/
```

---

## Phase 1: Core Infrastructure (v1.1.0)

**Goal:** Solidify the foundation for production use.

### 1.1 Traefik Integration
**Priority:** High
**Effort:** 2-3 days

Generate Traefik dynamic configuration automatically:
```bash
sl-cli site domain add my-site --domain mysite.com --traefik
# Generates Traefik dynamic config for automatic SSL via Let's Encrypt
```

**Deliverables:**
- CLI command: `sl-cli site domain add/remove/list`
- Traefik dynamic config generation (YAML format)
- Automatic Traefik hot-reload trigger
- Documentation for Traefik setup

### 1.2 Cloudflare Integration
**Priority:** High
**Effort:** 2-3 days

Manage Cloudflare DNS and SSL automatically:
```bash
sl-cli site domain add my-site --domain mysite.com --cloudflare
# Adds DNS record, enables SSL, configures CDN
```

**Deliverables:**
- CLI command: `sl-cli site domain cloudflare`
- Cloudflare API integration
- DNS record management (A, CNAME)
- SSL/TLS mode configuration
- Page rule setup (caching, security)

### 1.3 Domain Management
**Priority:** High
**Effort:** 1-2 days

Full domain CRUD operations:
```bash
sl-cli site domain add my-site --domain mysite.com
sl-cli site domain list my-site
sl-cli site domain remove my-site --domain mysite.com
```

**Deliverables:**
- Database schema for domains
- Domain repository
- CLI commands for domain management
- Validation and conflict detection

### 1.4 EJS Rendering
**Priority:** Medium
**Effort:** 2-3 days

Server-side template rendering with EJS:
```bash
# EJS templates supported
sl-cli site write my-site v1 "index.html" --content '<h1><%= title %></h1>'
sl-cli site write my-site v1 "index.html.data.json" --content '{"title":"My Site"}'
```

**Deliverables:**
- EJS template engine integration
- Data file support (`.html.data.json`)
- Template rendering at serve time
- Caching for performance

---

## Phase 2: Content Management (v1.2.0)

**Goal:** Enable richer content management.

### 2.1 Blog Module
**Priority:** High
**Effort:** 3-4 days

Blog functionality with posts, categories, and RSS:
```bash
sl-cli blog create my-blog --site my-site
sl-cli blog post create my-blog --title "Hello World" --content "..."
sl-cli blog post list my-blog
```

**Deliverables:**
- Blog database schema
- Blog CLI commands
- Post CRUD operations
- Category management
- RSS feed generation
- Blog routing (`/blog/`, `/blog/:slug`)

### 2.2 Asset Management
**Priority:** Medium
**Effort:** 2-3 days

Upload and manage static assets:
```bash
sl-cli site asset upload my-site v1 --file logo.png
sl-cli site asset list my-site v1
sl-cli site asset delete my-site v1 logo.png
```

**Deliverables:**
- Asset storage in file system
- Asset CLI commands
- Asset serving at `/assets/`
- Image optimization (optional)
- Asset versioning

### 2.3 SEO Meta Tags
**Priority:** Medium
**Effort:** 1-2 days

Per-page SEO configuration:
```bash
sl-cli site write my-site v1 "index.html" --seo-title "My Site" --seo-description "..."
```

**Deliverables:**
- SEO metadata in database
- CLI flags for SEO
- Automatic meta tag generation
- Open Graph tags
- Twitter Card tags

---

## Phase 3: Multi-Tenancy (v1.3.0)

**Goal:** Support multiple users and organizations.

### 3.1 User Management
**Priority:** Medium
**Effort:** 3-4 days

User authentication and authorization:
```bash
sl-cli user create --email user@example.com --password secret
sl-cli user list
sl-cli user set-role user@example.com --role admin
```

**Deliverables:**
- User database schema
- Password hashing (bcrypt)
- User CLI commands
- Role-based access control (admin, editor, viewer)
- JWT token generation

### 3.2 Organization Management
**Priority:** Medium
**Effort:** 2-3 days

Organization-based multi-tenancy:
```bash
sl-cli organization create --name "My Company"
sl-cli organization user add my-org user@example.com
sl-cli site create --name "Site" --slug "site" --org my-org
```

**Deliverables:**
- Organization database schema
- Organization CLI commands
- User-organization relationships
- Site-organization associations
- Permission checks

### 3.3 Authentication Middleware
**Priority:** High
**Effort:** 2-3 days

HTTP authentication for admin APIs:
```bash
# Protected endpoints require JWT token
curl -H "Authorization: Bearer <token>" http://localhost:3099/api/sites
```

**Deliverables:**
- JWT middleware
- Login endpoint
- Token refresh
- Protected route decorators
- Session management

---

## Phase 4: Advanced Features (v1.4.0)

**Goal:** Add power-user features.

### 4.1 Import/Export
**Priority:** Low
**Effort:** 2-3 days

Backup and restore functionality:
```bash
sl-cli site export my-site --output my-site.tar.gz
sl-cli site import my-site --input my-site.tar.gz
```

**Deliverables:**
- Export command (tar.gz with DB + files)
- Import command (validate and restore)
- Version selection for export
- Incremental export option

### 4.2 CLI JSON Mode
**Priority:** High
**Effort:** 1-2 days

Consistent JSON output for all commands:
```bash
sl-cli site list --json
# → {"sites":[{"id":"...","name":"...","slug":"..."}]}
```

**Deliverables:**
- `--json` flag for all commands
- Structured JSON responses
- Error envelopes in JSON
- Documentation for JSON schema

### 4.3 Webhooks
**Priority:** Low
**Effort:** 2-3 days

Webhook notifications for events:
```bash
sl-cli site webhook add my-site --url https://example.com/webhook --events create,update,delete
```

**Deliverables:**
- Webhook database schema
- Webhook CLI commands
- Event system
- HTTP webhook delivery
- Retry logic with exponential backoff

---

## Phase 5: Performance & Reliability (v1.5.0)

**Goal:** Production-ready performance.

### 5.1 Caching Layer
**Priority:** Medium
**Effort:** 2-3 days

In-memory caching for frequently accessed content:
```bash
sl-cli backend start --daemon --cache-enabled --cache-ttl 300
```

**Deliverables:**
- In-memory cache (LRU)
- Cache invalidation on version switch
- Cache statistics
- CLI flags for cache control

### 5.2 Rate Limiting
**Priority:** Medium
**Effort:** 1-2 days

Protect against abuse:
```bash
sl-cli backend start --daemon --rate-limit 100 --rate-window 60
```

**Deliverables:**
- Rate limiting middleware
- Per-IP rate limits
- Configurable limits
- Rate limit headers

### 5.3 Health Checks
**Priority:** High
**Effort:** 1 day

Monitoring and health endpoints:
```bash
curl http://localhost:3099/health
# → {"status":"healthy","version":"1.5.0","uptime":1234}
```

**Deliverables:**
- `/health` endpoint
- `/metrics` endpoint (Prometheus format)
- Database connectivity checks
- File system checks
- Uptime tracking

---

## Phase 6: Developer Experience (v1.6.0)

**Goal:** Make it easier for developers to use.

### 6.1 Watch Mode
**Priority:** Medium
**Effort**: 2-3 days

Auto-reload on file changes:
```bash
sl-cli backend start --daemon --watch
# Auto-reloads when files in ~/.superlandings/sites/ change
```

**Deliverables:**
- File watcher (fsnotify)
- Auto-reload on file changes
- Graceful reload (no downtime)
- Watch configuration

### 6.2 Development Server
**Priority:** Medium
**Effort**: 1-2 days

Hot reload for development:
```bash
sl-cli backend dev --port 3099
# Serves from current directory, auto-reloads
```

**Deliverables:**
- Development mode
- Hot reload
- Error pages
- Debug logging

### 6.3 CLI Autocomplete
**Priority:** Low
**Effort**: 1-2 days

Shell autocomplete for commands:
```bash
sl-cli site <TAB>
# → create, list, version, write
```

**Deliverables:**
- Shell completion scripts (bash, zsh, fish)
- Dynamic completion based on context
- Installation instructions

---

## Phase 7: Ecosystem (v2.0.0)

**Goal:** Build a plugin ecosystem.

### 7.1 Plugin System
**Priority:** Low
**Effort**: 4-5 days

Extensible plugin architecture:
```bash
sl-cli plugin install analytics
sl-cli plugin enable analytics my-site
```

**Deliverables:**
- Plugin interface
- Plugin discovery
- Plugin lifecycle (install, enable, disable)
- Plugin API (hooks, events)

### 7.2 Official Plugins
**Priority:** Low
**Effort**: 3-4 days per plugin

Core plugins:
- **Analytics** - Google Analytics, Plausible integration
- **Forms** - Form submission handling
- **Comments** - Comment system integration
- **Search** - Full-text search
- **Sitemap** - XML sitemap generation

---

## Future Considerations

### Potential Features (Not Scheduled)
- **Admin UI** - Web interface for site management (React embedded via go:embed)
- **Database Backups** - Automated SQLite backups
- **Multi-region Deployment** - Sync across multiple servers
- **GraphQL API** - GraphQL endpoint for advanced queries
- **WebAssembly Templates** - WASM for template rendering
- **Edge Deployment** - Support for edge platforms (Cloudflare Workers, Deno Deploy)

### Architecture Decisions to Revisit
- **SQLite vs PostgreSQL** - For high-traffic multi-tenant deployments
- **Single Binary vs Microservices** - For scaling beyond single server
- **File System vs S3** - For cloud-native storage

---

## Timeline Estimate

| Phase | Version | Estimated Time |
|-------|---------|----------------|
| Phase 1: Core Infrastructure | v1.1.0 | 1-2 weeks |
| Phase 2: Content Management | v1.2.0 | 1-2 weeks |
| Phase 3: Multi-Tenancy | v1.3.0 | 1-2 weeks |
| Phase 4: Advanced Features | v1.4.0 | 1-2 weeks |
| Phase 5: Performance & Reliability | v1.5.0 | 1 week |
| Phase 6: Developer Experience | v1.6.0 | 1 week |
| Phase 7: Ecosystem | v2.0.0 | 2-3 weeks |

**Total to v2.0.0:** ~8-12 weeks

**Note:** This is a rough estimate. Actual timeline depends on priorities, testing, and feedback.

---

## Contributing

We welcome contributions! See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

### How to Contribute

1. Check this roadmap for open tasks
2. Open an issue to discuss the feature
3. Submit a PR with tests
4. Ensure all tests pass
5. Update documentation

### Priority Guidelines

- **High Priority** - Core functionality, security, performance
- **Medium Priority** - Developer experience, content management
- **Low Priority** - Nice-to-have features, plugins

---

## Changelog

See [CHANGELOG.md](../CHANGELOG.md) for release history.

---

*This roadmap is a living document and may change based on community feedback and emerging needs.*