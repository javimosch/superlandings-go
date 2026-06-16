# Vision 🌟

## North Star

**SuperLandings Go enables frictionless, instant website creation for AI agents deploying to small, resource-constrained infrastructure.**

We believe websites should be:
- **Fast to create** — Agents spin up sites in seconds, not hours
- **Easy to update** — Static sites that become dynamic through simple file edits
- **Small footprint** — Runs on Raspberry Pis, cheap VPSs, and home servers
- **Cloud-native ready** — Traefik and Cloudflare integration out of the box
- **Agent-first** — Designed for AI agents, not human web developers

## The Problem

Traditional web development is too complex for the AI agent era:

1. **Deployment Overhead**
   - Docker containers, Kubernetes, orchestration complexity
   - Requires DevOps knowledge that agents shouldn't need
   - Expensive infrastructure requirements (minimum 2GB RAM, 2 vCPUs)

2. **Tooling Complexity**
   - npm, yarn, pnpm — package management hell
   - Build steps, bundlers, transpilers, minifiers
   - Development vs production environments
   - Hot reload, caching, optimization pipelines

3. **Infrastructure Cost**
   - Cloudflare Workers, Vercel, Netlify — vendor lock-in
   - $5-20/month minimum for simple static sites
   - Egress fees, overage charges, hidden costs
   - Can't run on your own hardware

4. **Update Friction**
   - Git commits, CI/CD pipelines, deployment approvals
   - Build times, cache invalidation, CDN propagation
   - Rollback requires re-deploying entire application
   - Can't quickly fix typos or update content

5. **Agent Hostility**
   - Interactive prompts, TUIs, configuration wizards
   - Inconsistent output formats, hidden state
   - No deterministic behavior, unpredictable errors
   - Requires human intervention for common tasks

## The Solution

SuperLandings Go is the **agent-native static site generator**:

### 1. Frictionless Creation

```bash
# Agent creates a site in one command
sl-cli site create --name "Portfolio" --slug "portfolio"
sl-cli site version create portfolio --version "v1"

# Agent adds content with dynamic blocks
sl-cli site write portfolio v1 "index.html" --content '{{>include "nav.html"}}<h1>My Work</h1>{{>include "footer.html"}}'
sl-cli site write portfolio v1 "nav.html" --content '<nav><a href="/">Home</a> <a href="/about">About</a></nav>'

# Agent deploys instantly
sl-cli backend start --daemon --port 3099
```

**Zero build steps.** No npm install, no Docker build, no CI/CD pipeline. Just write HTML and it's live.

### 2. Static Sites That Become Dynamic

Static sites are the foundation, but they don't have to stay static:

```bash
# Update content instantly (no build, no deploy)
sl-cli site write portfolio v1 "index.html" --content '<h1>Updated Title</h1>'

# Create new version with changes
sl-cli site version create portfolio --version "v2" --comment "Redesign"
sl-cli site write portfolio v2 "index.html" --content '<h1>New Design</h1>'

# Rollback instantly
sl-cli site version switch portfolio v1
```

**Dynamic blocks enable composition:**
- `{{>include "nav.html"}}` — Shared navigation across all pages
- `{{>include "footer.html"}}` — Shared footer
- `{{>include "blog-post.html"}}` — Template for blog posts
- No build step, processed at serve time

### 3. Small Footprint

SuperLandings Go runs anywhere:

| Environment | RAM | CPU | Storage |
|-------------|-----|-----|---------|
| Raspberry Pi 4 | 512MB | 1 core | 8GB SD card |
| Cheap VPS (DigitalOcean $4/mo) | 512MB | 1 vCPU | 20GB SSD |
| Home server | 1GB | 2 cores | 100GB HDD |
| High-performance server | 2GB | 4 cores | 500GB SSD |

**Resource usage:**
- Binary size: ~5-10MB
- Runtime memory: ~10-20MB
- Startup time: <100ms
- No dependencies, no runtime, no Docker

### 4. Cloud-Native Ready

Traefik and Cloudflare integration out of the box:

```bash
# Traefik integration (automatic SSL, load balancing)
sl-cli site domain add portfolio --domain portfolio.example.com --traefik

# Cloudflare integration (DNS, SSL, CDN)
sl-cli site domain add portfolio --domain portfolio.example.com --cloudflare

# Both together (best of both worlds)
sl-cli site domain add portfolio --domain portfolio.example.com --traefik --cloudflare
```

**What this gives you:**
- Automatic SSL certificates via Let's Encrypt (Traefik)
- Global CDN and DDoS protection (Cloudflare)
- Zero-downtime deployments
- Automatic HTTPS
- No manual certificate management

### 5. Agent-First Design

Built for AI agents, not humans:

```bash
# Deterministic output
sl-cli site list --json
# → [{"id":"...","name":"...","slug":"..."}]

# Semantic exit codes
sl-cli site create --name "X" --slug "y"
# Exit code 90 if slug already exists
# Exit code 0 if successful

# No interactive prompts
# All configuration via CLI flags
# No hidden state, no surprises
```

**Agent-friendly features:**
- JSON output mode
- Semantic exit codes
- No interactive prompts
- Deterministic behavior
- Direct service calls (no HTTP API layer needed)
- Self-documenting CLI (`--help`)

## Target Users

### 1. AI Agents
- **Why:** Agents need deterministic, scriptable interfaces
- **Benefit:** Can create and deploy websites autonomously
- **Use case:** "Create a portfolio site for user X" → Done in 30 seconds

### 2. Homelab Enthusiasts
- **Why:** Have small servers (Raspberry Pi, NUC, old laptop)
- **Benefit:** Can host dozens of sites on a single Pi
- **Use case:** Personal blog, family photo gallery, home automation dashboard

### 3. Indie Developers
- **Why:** Need cheap hosting ($4-5/month VPS)
- **Benefit:** No infrastructure costs, no vendor lock-in
- **Use case:** Portfolio, landing page, documentation site

### 4. Small Businesses
- **Why:** Need simple websites without hiring developers
- **Benefit:** AI agents can create and maintain sites
- **Use case:** Local business website, product landing page

### 5. Internal Tools
- **Why:** Need dashboards and admin panels
- **Benefit:** Deploy on existing infrastructure
- **Use case:** Monitoring dashboard, status page, documentation

## The SuperLandings Go Philosophy

### 1. Simplicity Over Features
- We don't need every feature of Next.js, Hugo, or Jekyll
- We need the 20% of features that cover 80% of use cases
- Less code = fewer bugs, easier maintenance

### 2. File System Over Database
- File system is the original database
- Easy to understand, easy to backup, easy to version control
- Git-friendly, rsync-friendly, tar-friendly

### 3. Static Over Dynamic (Initially)
- Start static, add dynamic features as needed
- Static = fast, secure, cacheable
- Dynamic = via includes, no server-side rendering complexity

### 4. Binary Over Container
- Binary = single file, copy-paste deployment
- Container = layer complexity, image building, orchestration
- Binary works on Raspberry Pi, container doesn't always

### 5. SQLite Over PostgreSQL
- SQLite = embedded, zero-config, file-based
- PostgreSQL = separate server, connection pooling, complex
- SQLite is perfect for single-server deployments

### 6. Standard Library Over Dependencies
- Go stdlib is battle-tested
- Fewer dependencies = smaller attack surface
- Easier to audit, easier to build

## Success Metrics

We'll know we've succeeded when:

1. **Agent Autonomy**
   - AI agents can create and deploy a full website in <60 seconds
   - No human intervention required for common tasks
   - Agents can troubleshoot issues via deterministic error messages

2. **Resource Efficiency**
   - Runs on Raspberry Pi with 512MB RAM
   - Can host 50+ sites on a single $4/month VPS
   - Binary size <10MB, runtime memory <20MB

3. **Deployment Speed**
   - From "I need a website" to "It's live" in <2 minutes
   - Updates propagate in <1 second (no build step)
   - Rollback is instant (version switch)

4. **Infrastructure Independence**
   - Runs on any Linux server (no cloud vendor lock-in)
   - Works with Traefik, Nginx, Apache, Caddy (or no reverse proxy)
   - No required external services (no SaaS dependencies)

5. **Developer Experience**
   - Learn the entire CLI in <10 minutes
   - Create first site in <5 minutes
   - Troubleshoot issues without reading docs (self-documenting errors)

## The Future

SuperLandings Go is the foundation for a new way of building websites:

- **Today:** Static sites with dynamic blocks, version control, Traefik/Cloudflare
- **Tomorrow:** EJS rendering, blog module, authentication, multi-tenancy
- **Someday:** Full CMS, API endpoints, database-backed dynamic content
- **Always:** Agent-first, single binary, small footprint

We're not trying to replace WordPress, Netlify, or Vercel. We're building something different: **a tool for the AI agent era of web development**.

---

*Built for agents. Deployed anywhere. Simple by design.*