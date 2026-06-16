---
name: dk2-deployment
description: Deploy SuperLandings Go to dk2 VM (92.113.145.16) with hotify-cli and Traefik
---

# SuperLandings Go Deployment to dk2

## Overview

Deploy SuperLandings Go to dk2 VM (92.113.145.16) with hotify-cli DNS management and Traefik reverse proxy.

## dk2 VM Details

- **Host:** 92.113.145.16
- **OS:** Ubuntu (LXC container)
- **User:** dk2 (UID 1002)
- **SSH key:** `~/.ssh/id_rsa_srv`

## Deployment Steps

### 1. Build sl-cli Binary

```bash
cd ~/ai/superlandings-go
go build -o sl-cli ./cmd/sl-cli
```

### 2. Deploy Binary to dk2

```bash
# Copy to /tmp first (permission workaround)
scp -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv sl-cli dk2:/tmp/

# Move to /usr/local/bin
ssh -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv dk2 "sudo mv /tmp/sl-cli /usr/local/bin/ && sudo chmod +x /usr/local/bin/sl-cli"
```

### 3. Create Config on dk2

```bash
ssh -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv dk2 "mkdir -p /home/dk2/.superlandings && echo '{\"sites_dir\": \"/home/dk2/.superlandings/sites\"}' > /home/dk2/.superlandings/config.json"
```

**Important:** Use `/home/dk2/.superlandings/` not `/root/.superlandings/` because dk2 user doesn't have root access.

### 4. Start Daemon on dk2

```bash
ssh -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv dk2 "sl-cli backend start --daemon --port 3100"
```

**Port notes:**
- Default port 3099 may be in use (Docker)
- Use port 3100 to avoid conflicts
- Daemon logs: `/home/dk2/.superlandings/sl-cli.log`

### 5. Sync Site Files

**Option A: Using sl-cli sync command**

```bash
sl-cli site sync slv2 --host 92.113.145.16 --user root --key ~/.ssh/id_rsa_srv
```

**Option B: Manual sync (current workaround)**

```bash
# Create site directory on dk2
ssh -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv dk2 "mkdir -p /home/dk2/.superlandings/sites/slv2/v1"

# Copy site files
scp -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv -r ~/.superlandings/sites/slv2/* dk2:/home/dk2/.superlandings/sites/slv2/

# Create version on dk2
ssh -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv dk2 "sl-cli site version create slv2 --version v1 --comment 'Synced from local'"
```

### 6. Configure DNS via hotify-cli

```bash
# From local machine (where hotify-cli is configured)
hotify-cli setup-dns --id slv2 --ip 92.113.145.16 --local
```

### 7. Configure Traefik on dk2

```bash
# Fix Traefik config ownership
ssh -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv dk2 "sudo chown dk2:dk2 /etc/traefik/dynamic.yml /etc/traefik/traefik.yml"

# Add slv2 routing config
ssh -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv dk2 "sudo tee /etc/traefik/dynamic.yml > /dev/null << 'EOF'
http:
  routers:
    rcmd:
      rule: \"Host(\`rcmd.intrane.fr\`)\"
      service: rcmd
      entryPoints:
        - websecure
      tls:
        certResolver: letsencrypt
        domains:
          - main: rcmd.intrane.fr

    slv2:
      rule: \"Host(\`slv2.intrane.fr\`)\"
      service: slv2
      entryPoints:
        - web
      middlewares:
        - slv2-addprefix

  services:
    rcmd:
      loadBalancer:
        servers:
          - url: \"http://127.0.0.1:3032\"

    slv2:
      loadBalancer:
        servers:
          - url: \"http://127.0.0.1:3100\"

  middlewares:
    slv2-addprefix:
      addPrefix:
        prefix: \"/slv2\"
EOF
"

# Restart Traefik
ssh -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv dk2 "sudo systemctl restart traefik"
```

### 8. Test Deployment

```bash
# Test local access on dk2
ssh -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv dk2 "curl -s http://localhost:3100/slv2/"

# Test domain access
curl -s http://slv2.intrane.fr
```

## Troubleshooting

### Port 3099 Already in Use

```bash
# Check what's using the port
ssh dk2 "sudo lsof -i :3099"

# Use port 3100 instead
sl-cli backend start --daemon --port 3100
```

### Traefik Permission Denied

```bash
# Fix config ownership
ssh dk2 "sudo chown dk2:dk2 /etc/traefik/dynamic.yml /etc/traefik/traefik.yml"
```

### Site Not Found (404)

```bash
# Check if site exists in database
ssh dk2 "sl-cli site list"
ssh dk2 "sl-cli site version list slv2"

# Check if files exist
ssh dk2 "ls -la /home/dk2/.superlandings/sites/slv2/v1/"

# Verify config points to correct directory
ssh dk2 "cat /home/dk2/.superlandings/config.json"
```

### Traefik Not Routing

```bash
# Check Traefik status
ssh dk2 "sudo systemctl status traefik"

# Check Traefik logs
ssh dk2 "sudo journalctl -u traefik -n 50 --no-pager"

# Verify dynamic config
ssh dk2 "sudo cat /etc/traefik/dynamic.yml"
```

## Current Deployment Status

**Live site:** http://slv2.intrane.fr

- ✅ sl-cli deployed to `/usr/local/bin/sl-cli`
- ✅ Daemon running on port 3100
- ✅ slv2 site synced and active
- ✅ DNS configured: slv2.intrane.fr → 92.113.145.16
- ✅ Traefik configured with path prefix middleware
- ✅ Domain accessible

## References

- hotify-integration skill for hotify-cli details
- jar-dk2-manage skill for dk2 VM management
- hotify-setup skill for general hotify-cli usage