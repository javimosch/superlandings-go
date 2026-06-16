---
name: remote-deployment
description: Deploy SuperLandings Go to remote servers with hotify-cli and Traefik
---

# SuperLandings Go Remote Deployment

## Overview

Deploy SuperLandings Go to remote servers with hotify-cli DNS management and Traefik reverse proxy.

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
ssh -i <SSH_KEY> <USER>@<SERVER_IP> "sl-cli backend start --daemon --port <PORT>"
```

**Port notes:**
- Default port 3099 may be in use
- Check available ports with `ss -tlnp | grep <PORT>`
- Use an available port to avoid conflicts
- Daemon logs: `/home/<USER>/.superlandings/sl-cli.log`

### 5. Sync Site Files

**Option A: Using sl-cli sync command**

```bash
sl-cli site sync <SITE_SLUG> --host <SERVER_IP> --user <USER> --key <SSH_KEY>
```

**Option B: Manual sync (current workaround)**

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
```

## References

- hotify-integration skill for hotify-cli details
- sync-mechanism skill for sync commands details
- hotify-setup skill for general hotify-cli usage