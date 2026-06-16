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
- It DOES generate router/service configuration in `/etc/traefik/dynamic.yml` (as of v2.10.1)
- Path prefix middleware is now supported via `--path-prefix` flag

## Deployment Workflow

### 1. Register App with hotify-cli

```bash
hotify-cli setup --id <APP_ID> --name <APP_NAME> --domain <DOMAIN> --port <PORT> --backend-url http://127.0.0.1:<PORT> --cmd 'sleep infinity' --path-prefix /<SITE_SLUG>
```

**Why `sleep infinity`?**
- hotify-cli expects a command that keeps the process alive
- `true` exits immediately, breaking the proxy
- `sleep infinity` keeps the process running as a placeholder

**Why `--path-prefix`?**
- Services like sl-cli serve sites at paths like `/<SITE_SLUG>/`, `/other-site/`
- The `--path-prefix` flag automatically generates Traefik addPrefix middleware
- This routes domain root to the correct path without manual configuration

### 2. Configure DNS

```bash
hotify-cli setup-dns --id <APP_ID> --ip <SERVER_IP> --local
```

### 3. Fix Traefik Config Ownership (if needed)

```bash
# Traefik config must be owned by the user running hotify-cli
sudo chown <USER>:<USER> /etc/traefik/dynamic.yml /etc/traefik/traefik.yml
```

### 4. Setup Traefik (automatic with --path-prefix)

```bash
hotify-cli setup-traefik --id <APP_ID>
```

This automatically generates the router, service, and addPrefix middleware in `/etc/traefik/dynamic.yml` when `--path-prefix` was used during setup.

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

# Proxy setup (now supports --path-prefix)
sl-cli site proxy <SITE_SLUG> --domain <DOMAIN> --internal-url http://127.0.0.1:<PORT> --path-prefix /<SITE_SLUG>
```

## Known Issues

1. **Config ownership** - Traefik files must be owned by the user running hotify-cli
2. **Domain duplication** - hotify-cli may append base domain twice
3. **Placeholder command required** - Need `sleep infinity` instead of `true` for backend-url
4. **HTTPS connection reset** - Tailscale Funnel or iptables redirects may block port 443
   - Check: `tailscale funnel status` (disable with `tailscale funnel reset`)
   - Check: `iptables -t nat -L -n -v` (remove redirect: `iptables -t nat -D PREROUTING -p tcp --dport 443 -j REDIRECT --to-port 8443`)
   - See remote-deployment skill for detailed troubleshooting

**Fixed in hotify-cli v2.10.1:**
- ✅ Router/service configuration now generated automatically
- ✅ Path prefix middleware support via `--path-prefix` flag

## References

- hotify-setup skill for general hotify-cli usage
- dk2-deployment skill for server-specific deployment details
- GitHub issue: https://github.com/javimosch/hotify-cli/issues/1