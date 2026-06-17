# Vision 🌟

## North Star

**SuperLandings Go enables frictionless, instant website creation for AI agents deploying to small, resource-constrained infrastructure.**

## The Problem

Traditional web development is too complex for the AI agent era:

1. **Deployment Overhead** — Docker, Kubernetes, complex orchestration
2. **Tooling Complexity** — npm, build steps, bundlers, transpilers
3. **Infrastructure Cost** — Cloudflare Workers, Vercel, vendor lock-in
4. **Update Friction** — Git commits, CI/CD pipelines, build times
5. **Agent Hostility** — Interactive prompts, inconsistent output, hidden state

## The Solution

SuperLandings Go is the agent-native static site generator:

### 1. Frictionless Creation
```bash
sl-cli site create --name "Portfolio" --slug "portfolio"
sl-cli site version create portfolio --version "v1"
sl-cli site write portfolio v1 "index.html" --content '{{>include "nav.html"}}<h1>My Work</h1>{{>include "footer.html"}}'
sl-cli backend start --daemon --port 3099
```

Zero build steps. No npm install, no Docker build, no CI/CD pipeline.

### 2. Static Sites That Become Dynamic
```bash
# Update content instantly (no build, no deploy)
sl-cli site write portfolio v1 "index.html" --content '<h1>Updated Title</h1>'

# Create new version
sl-cli site version create portfolio --version "v2" --comment "Redesign"

# Rollback instantly
sl-cli site version switch portfolio v1
```

Dynamic blocks enable composition: `{{>include "nav.html"}}`, `{{>include "footer.html"}}`

### 3. Small Footprint
Runs anywhere: Raspberry Pi (512MB RAM), cheap VPS ($4/mo), home servers.

**Resource usage:**
- Binary size: ~5-10MB
- Runtime memory: ~10-20MB
- Startup time: <100ms
- No dependencies, no runtime, no Docker

### 4. Cloud-Native Ready
Traefik and Cloudflare integration out of the box:
```bash
sl-cli site domain add portfolio --domain portfolio.example.com --traefik --cloudflare
```

Automatic SSL certificates via Let's Encrypt, global CDN and DDoS protection via Cloudflare.

### 5. Agent-First Design
```bash
# Deterministic output
sl-cli site list --json
# → [{"id":"...","name":"...","slug":"..."}]

# Semantic exit codes
# Exit code 90 if slug already exists
# Exit code 0 if successful
```

No interactive prompts, all configuration via CLI flags, deterministic behavior.

## Target Users

- **AI Agents** — Deterministic, scriptable interfaces for autonomous website creation
- **Homelab Enthusiasts** — Host dozens of sites on a single Pi
- **Indie Developers** — Cheap hosting ($4-5/month VPS), no infrastructure costs
- **Small Businesses** — AI agents create and maintain sites
- **Internal Tools** — Dashboards, admin panels on existing infrastructure

## The SuperLandings Go Philosophy

1. **Simplicity Over Features** — 20% of features covering 80% of use cases
2. **File System Over Database** — Easy to understand, backup, version control
3. **Static Over Dynamic** — Start static, add dynamic features as needed
4. **Binary Over Container** — Single file deployment vs container complexity
5. **SQLite Over PostgreSQL** — Embedded, zero-config vs database server
6. **Standard Library Over Dependencies** — Battle-tested stdlib vs external deps

## Success Metrics

1. **Agent Autonomy** — AI agents create and deploy full website in <60 seconds
2. **Resource Efficiency** — 50+ sites on a single $4/month VPS
3. **Deployment Speed** — <2 minutes from idea to live
4. **Infrastructure Independence** — Runs on any Linux server
5. **Developer Experience** — Learn CLI in <10 minutes, create first site in <5 minutes

## The Future

- **Today:** Static sites with dynamic blocks, Go templates, version control, Traefik/Cloudflare
- **Tomorrow:** Blog module, authentication, multi-tenancy
- **Someday:** Full CMS, API endpoints, database-backed dynamic content
- **Always:** Agent-first, single binary, small footprint

We're not trying to replace WordPress, Netlify, or Vercel. We're building something different: **a tool for the AI agent era of web development**.

---

*Built for agents. Deployed anywhere. Simple by design.*