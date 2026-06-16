---
name: remote-deployment
description: Deploy SuperLandings Go to remote servers with hotify-cli and Traefik
---

# SuperLandings Go Remote Deployment

## Overview

Deploy SuperLandings Go to remote servers with hotify-cli DNS management and Traefik reverse proxy.

## New: Remote Execution via HTTP API

The sl-cli now supports remote execution via HTTP API, allowing you to manage sites from local CLI without direct SSH access to the server.

### Remote Execution Setup

```bash
# 1. Add remote target (one-time setup)
sl-cli targets add --name dk2 --host 92.113.145.16 --port 3100 --token <auth-token> --default

# 2. Start daemon with auth token on remote server
ssh user@server "sl-cli backend start --daemon --port 3100 --auth-token <auth-token>"

# 3. Use remote commands from local (no SSH needed!)
sl-cli site list --target dk2
sl-cli backend status --target dk2
sl-cli site sync slv2 --target dk2
```

### Remote Sync Configuration

For the daemon to perform syncs, configure a sync target on the remote server:

```bash
# Start daemon with sync target configured
sl-cli backend start --daemon --port 3100 --auth-token <auth-token> \
  --sync-host <production-server> \
  --sync-user root \
  --sync-key ~/.ssh/production_key
```

The sync will:
1. Rsync site files from daemon to production server
2. Export/import site metadata via sl-cli
3. Restart daemon on production server

## Prerequisites

- SSH access to remote server
- SSH key configured for passwordless auth
- hotify-cli installed on local machine
- Traefik installed and running on remote server
- Sudo access on remote server (for Traefik config)

## Deployment Steps

### 1. Build sl-cli Binary

```bash
cd ~/ai/superlandings-go
go build -o sl-cli ./cmd/sl-cli
```

### 2. Deploy Binary to Remote Server

```bash
# Copy to /tmp first (permission workaround)
scp -i <SSH_KEY> sl-cli <USER>@<SERVER_IP>:/tmp/

# Move to /usr/local/bin
ssh -i <SSH_KEY> <USER>@<SERVER_IP> "sudo mv /tmp/sl-cli /usr/local/bin/ && sudo chmod +x /usr/local/bin/sl-cli"
```

### 3. Create Config on Remote Server

```bash
ssh -i <SSH_KEY> <USER>@<SERVER_IP> "mkdir -p /home/<USER>/.superlandings && echo '{\"sites_dir\": \"/home/<USER>/.superlandings/sites\"}' > /home/<USER>/.superlandings/config.json"
```

**Important:** Use the user's home directory, not `/root/`, to avoid permission issues.

### 4. Start Daemon on Remote Server

```bash
# With authentication and sync target (recommended)
ssh -i <SSH_KEY> <USER>@<SERVER_IP> "sl-cli backend start --daemon --port <PORT> --auth-token <TOKEN> --sync-host <PROD_SERVER> --sync-key ~/.ssh/prod_key"

# Simple start (no auth, no sync)
ssh -i <SSH_KEY> <USER>@<SERVER_IP> "sl-cli backend start --daemon --port <PORT>"
```

**Port notes:**
- Default port 3099 may be in use
- Check available ports with `ss -tlnp | grep <PORT>`
- Use an available port to avoid conflicts
- Daemon logs: `/home/<USER>/.superlandings/sl-cli.log`

### 5. Sync Site Files

**Option A: Using remote execution (NEW - recommended)**

```bash
# Setup target once
sl-cli targets add --name dk2 --host 92.113.145.16 --port 3100 --token <auth-token>

# Sync via HTTP API
sl-cli site sync <SITE_SLUG> --target dk2
```

**Option B: Using sl-cli sync command (SSH-based)**

```bash
sl-cli site sync <SITE_SLUG> --host <SERVER_IP> --user <USER> --key <SSH_KEY>
```

**Option C: Manual sync (legacy)**

```bash
# Create site directory on remote
ssh -i <SSH_KEY> <USER>@<SERVER_IP> "mkdir -p /home/<USER>/.superlandings/sites/<SITE_SLUG>/<VERSION>"

# Copy site files
scp -i <SSH_KEY> -r ~/.superlandings/sites/<SITE_SLUG>/* <USER>@<SERVER_IP>:/home/<USER>/.superlandings/sites/<SITE_SLUG>/

# Create version on remote
ssh -i <SSH_KEY> <USER>@<SERVER_IP> "sl-cli site version create <SITE_SLUG> --version <VERSION> --comment 'Synced from local'"
```

### 6. Configure DNS via hotify-cli

```bash
# From local machine (where hotify-cli is configured)
hotify-cli setup-dns --id <APP_ID> --ip <SERVER_IP> --local
```

### 7. Configure Traefik on Remote Server

```bash
# Fix Traefik config ownership
ssh -i <SSH_KEY> <USER>@<SERVER_IP> "sudo chown <USER>:<USER> /etc/traefik/dynamic.yml /etc/traefik/traefik.yml"

# Add routing config (preserve existing routes)
ssh -i <SSH_KEY> <USER>@<SERVER_IP> "sudo tee /etc/traefik/dynamic.yml > /dev/null << 'EOF'
http:
  routers:
    <APP_ID>:
      rule: \"Host(\`<DOMAIN>\`)\"
      service: <APP_ID>
      entryPoints:
        - web
      middlewares:
        - <APP_ID>-addprefix

  services:
    <APP_ID>:
      loadBalancer:
        servers:
          - url: \"http://127.0.0.1:<PORT>\"

  middlewares:
    <APP_ID>-addprefix:
      addPrefix:
        prefix: \"/<SITE_SLUG>\"
EOF
"

# Restart Traefik
ssh -i <SSH_KEY> <USER>@<SERVER_IP> "sudo systemctl restart traefik"
```

### 8. Test Deployment

```bash
# Test local access on remote server
ssh -i <SSH_KEY> <USER>@<SERVER_IP> "curl -s http://localhost:<PORT>/<SITE_SLUG>/"

# Test domain access
curl -s http://<DOMAIN>

# Test remote execution
sl-cli site list --target dk2
sl-cli backend status --target dk2
```

## Remote Execution Commands

### Target Management

```bash
# List configured targets
sl-cli targets list

# Add a new target
sl-cli targets add --name <name> --host <host> --port <port> --token <auth-token> --default

# Remove a target
sl-cli targets remove <name>
```

### Remote Operations

```bash
# List sites on remote
sl-cli site list --target <target-name>

# Check daemon status on remote
sl-cli backend status --target <target-name>

# Sync site to remote (via HTTP API)
sl-cli site sync <site-slug> --target <target-name>
```

## Troubleshooting

### Port Already in Use

```bash
# Check what's using the port
ssh <USER>@<SERVER_IP> "sudo lsof -i :<PORT>"

# Use a different port
sl-cli backend start --daemon --port <AVAILABLE_PORT>
```

### Traefik Permission Denied

```bash
# Fix config ownership
ssh <USER>@<SERVER_IP> "sudo chown <USER>:<USER> /etc/traefik/dynamic.yml /etc/traefik/traefik.yml"
```

### Site Not Found (404)

```bash
# Check if site exists in database
ssh <USER>@<SERVER_IP> "sl-cli site list"
ssh <USER>@<SERVER_IP> "sl-cli site version list <SITE_SLUG>"

# Check if files exist
ssh <USER>@<SERVER_IP> "ls -la /home/<USER>/.superlandings/sites/<SITE_SLUG>/<VERSION>/"

# Verify config points to correct directory
ssh <USER>@<SERVER_IP> "cat /home/<USER>/.superlandings/config.json"
```

### Traefik Not Routing

```bash
# Check Traefik status
ssh <USER>@<SERVER_IP> "sudo systemctl status traefik"

# Check Traefik logs
ssh <USER>@<SERVER_IP> "sudo journalctl -u traefik -n 50 --no-pager"

# Verify dynamic config
ssh <USER>@<SERVER_IP> "sudo cat /etc/traefik/dynamic.yml"

# Note: Traefik auto-reloads when config changes (watch: true in providers.file)
# Do not manually restart Traefik unless necessary
```

### HTTPS Connection Reset (port 443)

If HTTPS returns connection reset but HTTP works:

```bash
# Check for Tailscale Funnel (blocks port 443)
ssh <USER>@<SERVER_IP> "sudo tailscale funnel status"
# If enabled, disable it:
ssh <USER>@<SERVER_IP> "sudo tailscale funnel reset"

# Check for iptables redirect rules
ssh <USER>@<SERVER_IP> "sudo iptables -t nat -L -n -v | grep 443"
# If you see "tcp dpt:443 redir ports 8443", remove it:
ssh <USER>@<SERVER_IP> "sudo iptables -t nat -D PREROUTING -p tcp --dport 443 -j REDIRECT --to-port 8443"

# Save iptables rules permanently
ssh <USER>@<SERVER_IP> "sudo mkdir -p /etc/iptables && sudo iptables-save | sudo tee /etc/iptables/rules.v4 > /dev/null"

# Restart Tailscale if needed
ssh <USER>@<SERVER_IP> "sudo tailscale up"
```

**Root causes:**
- Tailscale Funnel uses port 443 and conflicts with Traefik HTTPS
- Old iptables redirect rules (443 → 8443) from previous configurations
- These rules block external HTTPS access while local access works fine

### Remote Execution Fails

```bash
# Check if daemon is running with auth token
ssh <USER>@<SERVER_IP> "ps aux | grep sl-cli"

# Check daemon logs
ssh <USER>@<SERVER_IP> "tail -50 /home/<USER>/.superlandings/sl-cli.log"

# Verify target configuration
sl-cli targets list

# Test API endpoint directly
curl -H "Authorization: Bearer <token>" http://<host>:<port>/api/status
```

### Sync Target Not Configured

```bash
# Check if sync target is configured on daemon
curl -H "Authorization: Bearer <token>" -X POST http://<host>:<port>/api/sites/<slug>/sync

# If returns "sync target not configured on daemon", restart daemon with sync flags
sl-cli backend start --daemon --port <port> --auth-token <token> \
  --sync-host <prod-server> --sync-key ~/.ssh/prod_key
```

## Learnings & Caveats

### Remote Execution Limitations

1. **HTTP API sync returns plain text errors**: The remote sync endpoint via HTTP API (`--target`) currently returns plain text error messages instead of JSON. Use SSH-based sync (`--host`) for reliable error reporting until this is fixed.

2. **Sync path bug fixed**: The sync service was missing the home directory (`~`) in remote paths, causing "No such file or directory" errors. This has been fixed - remote path is now `~/.superlandings/sites/<slug>` instead of `/.superlandings/sites/<slug>`.

3. **Port support added**: SSH, SCP, and rsync commands now properly support non-standard SSH ports via `-p` flag for SSH and `-P` flag for SCP.

4. **Don't restart daemon unnecessarily**: After sync completes, files are immediately live. The daemon restart step at the end of sync is optional - only restart if you need to reload configuration or if the daemon crashed.

5. **SSH-based sync is more reliable**: For now, use SSH-based sync (`--host --user --key`) instead of HTTP API sync (`--target`) until the JSON error response issue is resolved.

6. **Daemon restart removed from sync**: The sync service no longer restarts the daemon after syncing. This was unnecessary because the daemon reads files on each request. Sync now:
   - Rsyncs site files to remote
   - Copies and imports metadata
   - Returns success immediately
   - Changes are live without any restart
   - No more exit 255 failures

7. **Real-time sync without restart**: Files are served immediately after sync - daemon reads from disk on each request, no restart needed. **Confirmed test:**
```bash
# Production workflow (use sl-cli sync):
sl-cli site sync <slug> --host <server> --user <user> --key <key>
# Changes appear immediately, no restart needed
```
**Result:** Changes appear instantly without daemon restart. The daemon restart step has been removed from sync service entirely.

8. **Dynamic blocks ARE implemented**: The `{{>include "path"}}` syntax works. Implemented in `internal/services/site.go` (lines 277-301) with:
   - Regex pattern matching for `{{>include "path"}}`
   - Recursive nested includes support
   - File reading from version directory
   - **Tested and working** on both local and remote
```bash
# Create nav.html and footer.html in version directory
# Use in index.html: {{>include "nav.html"}}
# Daemon processes includes on each request
```

### Deployment Best Practices

1. **Verify files before daemon restart**: After sync, check if files exist on remote (`ls -la ~/.superlandings/sites/<slug>/<version>/`) before restarting daemon.

2. **Test site access directly**: Use `curl http://<server>:<port>/<slug>/` to test site access instead of assuming daemon restart is needed.

3. **Keep daemon alive during sync**: The sync process transfers files and imports metadata while daemon is running - no need to stop it first.

4. **Use SSH for initial deployment**: SSH-based sync (`--host`) is more reliable for initial deployment. HTTP API sync (`--target`) is better for ongoing operations once daemon is stable.

5. **Check sync logs**: If sync fails, check the specific error output - rsync, scp, and ssh each have their own error messages that help diagnose issues.

## Migration from sl-cli v1 to v2

### Process to migrate from Node.js sl-cli to Go sl-cli:

1. **Export from v1:**
```bash
cd ~/ai/superlandings
MODE=staging node cli/index.js landing content get <slug> --output /tmp/<slug>.html
```

2. **Create site in v2:**
```bash
cd ~/ai/superlandings-go
sl-cli site create --name "<Site Name>" --slug <slug>
sl-cli site version create <slug> --version v1 --comment "Migrated from v1"
```

3. **Copy content:**
```bash
cp /tmp/<slug>.html ~/.superlandings/sites/<slug>/v1/index.html
```

4. **Deploy to remote:**
```bash
# Create site on remote first
ssh user@server "sl-cli site create --name '<Site Name>' --slug <slug>"
ssh user@server "mkdir -p ~/.superlandings/sites/<slug>/v1"

# Copy files
scp ~/.superlandings/sites/<slug>/v1/index.html user@server:~/.superlandings/sites/<slug>/v1/index.html

# Create version on remote
ssh user@server "sl-cli site version create <slug> --version v1 --comment 'Migrated from v1'"
```

5. **Configure Traefik on remote:**
```bash
# Add to /etc/traefik/dynamic.yml
# Include router, service, and middleware with addPrefix
sudo systemctl restart traefik
```

6. **Configure DNS:**
- Use hotify-cli or manually update Cloudflare
- Point domain to server IP
- Ensure Traefik SSL cert is configured

### Migration Example (intrane.fr → intrane.intrane.fr):

```bash
# Export from v1
cd ~/ai/superlandings
MODE=staging node cli/index.js landing content get intrane-fr --output /tmp/intrane-fr.html

# Create in v2
cd ~/ai/superlandings-go
sl-cli site create --name "Intrane.fr" --slug intrane
sl-cli site version create intrane --version v1 --comment "Migrated from sl-cli v1"
cp /tmp/intrane-fr.html ~/.superlandings/sites/intrane/v1/index.html

# Deploy to dk2
ssh dk2 "sl-cli site create --name 'Intrane.fr' --slug intrane"
ssh dk2 "mkdir -p ~/.superlandings/sites/intrane/v1"
scp ~/.superlandings/sites/intrane/v1/index.html dk2@server:~/.superlandings/sites/intrane/v1/index.html
ssh dk2 "sl-cli site version create intrane --version v1 --comment 'Migrated from v1'"

# Configure Traefik (add to /etc/traefik/dynamic.yml)
# Restart traefik
```

**Note:** Sync command (`sl-cli site sync`) may fail if site exists on remote but has no versions. Manual copy is more reliable for initial migration.

## References

- hotify-integration skill for hotify-cli details
- sync-mechanism skill for sync commands details
- hotify-setup skill for general hotify-cli usage