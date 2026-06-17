# Roadmap 🗺️

This roadmap outlines the planned development of SuperLandings Go.

## Current Status (v1.0.0)

### ✅ Completed
- Site management (create, list)
- Version control (file system based, instant rollback)
- Dynamic blocks (`{{>include "path"}}`)
- Go templates (variables, conditionals, loops)
- Hybrid storage (SQLite metadata + file system content)
- Sub-path routing (`/site/page`)
- HTTP server
- Daemon management
- Daemon auto-install (systemd)
- Agent-first CLI (JSON output, semantic exit codes)

---

## Phase 1: Core Infrastructure (v1.1.0)

**Estimated Time:** 1-2 weeks

### 1.1 Traefik Integration
Generate Traefik dynamic configuration automatically for SSL via Let's Encrypt.

### 1.2 Cloudflare Integration
Manage Cloudflare DNS and SSL automatically.

### 1.3 Domain Management
Full domain CRUD operations with validation and conflict detection.

### 1.4 Go Templates ✅
Native Go html/template with variables, conditionals, loops, and XSS protection.

---

## Phase 2: Content Management (v1.2.0)

**Estimated Time:** 1-2 weeks

### 2.1 Blog Module
Blog functionality with posts, categories, and RSS feeds.

### 2.2 Asset Management
Upload and manage static assets (images, CSS, JS).

### 2.3 SEO Meta Tags
Per-page SEO configuration with Open Graph and Twitter Card tags.

---

## Phase 3: Multi-Tenancy (v1.3.0)

**Estimated Time:** 1-2 weeks

### 3.1 User Management
User authentication and authorization with role-based access control.

### 3.2 Multi-tenancy
Per-site user access with role-based permissions.

### 3.3 Authentication Middleware
HTTP authentication for admin APIs using JWT tokens.

---

## Phase 4: Advanced Features (v1.4.0)

**Estimated Time:** 1-2 weeks

### 4.1 Import/Export
Backup and restore functionality with tar.gz export format.

### 4.2 CLI JSON Mode
Consistent JSON output for all commands.

### 4.3 Webhooks
Webhook notifications for create/update/delete events.

---

## Phase 5: Performance & Reliability (v1.5.0)

**Estimated Time:** 1 week

### 5.1 Caching Layer
In-memory LRU cache for frequently accessed content.

### 5.2 Rate Limiting
Per-IP rate limiting to protect against abuse.

### 5.3 Health Checks
Monitoring endpoints (`/health`, `/metrics`) for observability.

---

## Phase 6: Developer Experience (v1.6.0)

**Estimated Time:** 1 week

### 6.1 Watch Mode
Auto-reload on file changes for development.

### 6.2 Development Server
Hot reload development mode.

### 6.3 CLI Autocomplete

---

## Phase 7: Ecosystem (v2.0.0)

**Estimated Time:** 2-3 weeks

### 7.1 Plugin System
Extensible plugin architecture with hooks and events.

### 7.2 Official Plugins
Core plugins: analytics, forms, comments, search, sitemap.

---

## Timeline

| Phase | Version | Time |
|-------|---------|------|
| Phase 1: Core Infrastructure | v1.1.0 | 1-2 weeks |
| Phase 2: Content Management | v1.2.0 | 1-2 weeks |
| Phase 3: Multi-Tenancy | v1.3.0 | 1-2 weeks |
| Phase 4: Advanced Features | v1.4.0 | 1-2 weeks |
| Phase 5: Performance & Reliability | v1.5.0 | 1 week |
| Phase 6: Developer Experience | v1.6.0 | 1 week |
| Phase 7: Ecosystem | v2.0.0 | 2-3 weeks |

**Total to v2.0.0:** ~8-12 weeks

---

## Contributing

We welcome contributions! Check this roadmap for open tasks, open an issue to discuss features, and submit PRs with tests.

---

*This roadmap is a living document and may change based on community feedback.*