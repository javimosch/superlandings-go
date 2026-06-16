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
hotify-cli setup --id slv2 --name slv2 --domain slv2.intrane.fr --port 3100 --backend-url http://127.0.0.1:3100 --cmd 'sleep infinity'
```

**Why `sleep infinity`?**
- hotify-cli expects a command that keeps the process alive
- `true` exits immediately, breaking the proxy
- `sleep infinity` keeps the process running as a placeholder

### 2. Configure DNS

```bash
hotify-cli setup-dns --id slv2 --ip 92.113.145.16 --local
```

### 3. Fix Traefik Config Ownership (if needed)

```bash
# On dk2, Traefik config must be owned by dk2 user
sudo chown dk2:dk2 /etc/traefik/dynamic.yml /etc/traefik/traefik.yml
```

### 4. Manual Traefik Configuration

```bash
sudo tee /etc/traefik/dynamic.yml > /dev/null << 'EOF'
http:
  routers:
    slv2:
      rule: "Host(\`slv2.intrane.fr\`)"
      service: slv2
      entryPoints:
        - web
      middlewares:
        - slv2-addprefix
  services:
    slv2:
      loadBalancer:
        servers:
          - url: "http://127.0.0.1:3100"
  middlewares:
    slv2-addprefix:
      addPrefix:
        prefix: "/slv2"
EOF
```

**Why path prefix middleware?**
- sl-cli serves sites at paths like `/slv2/`, `/template-demo/`
- Domain root (`slv2.intrane.fr`) needs to be routed to `/slv2/`
- Traefik addPrefix middleware adds `/slv2` to the request path

### 5. Restart Traefik

```bash
sudo systemctl restart traefik
```

## Domain Duplication Issue

hotify-cli sometimes duplicates the domain suffix:

**Problem:** `slv2.intrane.fr.intrane.fr`

**Fix:**
```bash
sed -i 's|slv2.intrane.fr.intrane.fr|slv2.intrane.fr|g' ~/.hotify/config.json
```

## GitHub Issue

See https://github.com/javimosch/hotify-cli/issues/1 for the feature request to add router/service configuration generation to hotify-cli.

## sl-cli Commands

SuperLandings Go includes CLI commands that wrap hotify-cli:

```bash
# DNS setup
sl-cli site dns setup slv2 --domain slv2.intrane.fr
sl-cli site dns list slv2
sl-cli site dns remove slv2

# Proxy setup (currently requires manual Traefik config)
sl-cli site proxy slv2 --domain slv2.intrane.fr --internal-url http://127.0.0.1:3100
```

## Known Issues

1. **No automatic router/service config** - Must manually edit `/etc/traefik/dynamic.yml`
2. **Config ownership** - Traefik files must be owned by the user running hotify-cli
3. **Domain duplication** - hotify-cli may append base domain twice
4. **Placeholder command required** - Need `sleep infinity` instead of `true` for backend-url

## References

- hotify-setup skill for general hotify-cli usage
- jar-dk2-manage skill for dk2-specific Traefik configuration
- GitHub issue: https://github.com/javimosch/hotify-cli/issues/1