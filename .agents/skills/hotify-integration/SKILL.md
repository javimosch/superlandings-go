---
name: hotify-integration
description: hotify-cli integration for SuperLandings Go - DNS, Traefik, and reverse proxy configuration
---

# hotify-cli Integration for SuperLandings Go

## Overview

SuperLandings Go integrates with hotify-cli for DNS management and Traefik reverse proxy configuration. However, there are important limitations to understand.

## Critical Limitation: setup-traefik vs Router Configuration

**setup-traefik is for ACME certificates ONLY**

The `hotify-cli setup-traefik` command is specifically for ACME certificate setup (SSL/TLS), NOT for router/service configuration in Traefik.

When using `--backend-url`:
- hotify-cli stores the backend URL in `~/.hotify/config.json`
- But it does NOT generate router/service configuration in `/etc/traefik/dynamic.yml`
- Manual Traefik config is required for path prefix middleware

## Deployment Workflow

### 1. Register App with hotify-cli

```bash
hotify-cli setup --id <APP_ID> --name <APP_NAME> --domain <DOMAIN> --port <PORT> --backend-url http://127.0.0.1:<PORT> --cmd 'sleep infinity'
```

**Why `sleep infinity`?**
- hotify-cli expects a command that keeps the process alive
- `true` exits immediately, breaking the proxy
- `sleep infinity` keeps the process running as a placeholder

### 2. Configure DNS

```bash
hotify-cli setup-dns --id <APP_ID> --ip <SERVER_IP> --local
```

### 3. Fix Traefik Config Ownership (if needed)

```bash
# Traefik config must be owned by the user running hotify-cli
sudo chown <USER>:<USER> /etc/traefik/dynamic.yml /etc/traefik/traefik.yml
```

### 4. Manual Traefik Configuration

```bash
sudo tee /etc/traefik/dynamic.yml > /dev/null << 'EOF'
http:
  routers:
    <APP_ID>:
      rule: "Host(\`<DOMAIN>\`)"
      service: <APP_ID>
      entryPoints:
        - web
      middlewares:
        - <APP_ID>-addprefix
  services:
    <APP_ID>:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1:<PORT>"
  middlewares:
    <APP_ID>-addprefix:
      addPrefix:
        prefix: "/<SITE_SLUG>"
EOF
```

**Why path prefix middleware?**
- sl-cli serves sites at paths like `/<SITE_SLUG>/`, `/other-site/`
- Domain root needs to be routed to `/<SITE_SLUG>/`
- Traefik addPrefix middleware adds the site slug to the request path

### 5. Restart Traefik

```bash
sudo systemctl restart traefik
```

## Domain Duplication Issue

hotify-cli sometimes duplicates the domain suffix:

**Problem:** `example.com.example.com`

**Fix:**
```bash
sed -i 's|example.com.example.com|example.com|g' ~/.hotify/config.json
```

## GitHub Issue

See https://github.com/javimosch/hotify-cli/issues/1 for the feature request to add router/service configuration generation to hotify-cli.

## sl-cli Commands

SuperLandings Go includes CLI commands that wrap hotify-cli:

```bash
# DNS setup
sl-cli site dns setup <SITE_SLUG> --domain <DOMAIN>
sl-cli site dns list <SITE_SLUG>
sl-cli site dns remove <SITE_SLUG>

# Proxy setup (currently requires manual Traefik config)
sl-cli site proxy <SITE_SLUG> --domain <DOMAIN> --internal-url http://127.0.0.1:<PORT>
```

## Known Issues

1. **No automatic router/service config** - Must manually edit `/etc/traefik/dynamic.yml`
2. **Config ownership** - Traefik files must be owned by the user running hotify-cli
3. **Domain duplication** - hotify-cli may append base domain twice
4. **Placeholder command required** - Need `sleep infinity` instead of `true` for backend-url

## References

- hotify-setup skill for general hotify-cli usage
- dk2-deployment skill for server-specific deployment details
- GitHub issue: https://github.com/javimosch/hotify-cli/issues/1