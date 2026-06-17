---
name: remote-deployment
description: Deploy SuperLandings Go to remote servers with hotify-cli and Traefik
---

# SuperLandings Go Remote Deployment

## Remote Execution via HTTP API

Manage sites from local CLI via a remote daemon's HTTP API — no SSH needed for day-to-day ops.

```bash
# Add target (one-time)
sl-cli targets add --name <name> --host <IP> --port 3100 --token <token> --default

# Use --target on any command
sl-cli site list --target dk2
sl-cli backend status --target dk2
sl-cli site upload site "img.png" --file ./img.png --target dk2
```

## Initial Deployment

### 1. Build & Copy Binary

```bash
go build -o sl-cli ./cmd/sl-cli
scp sl-cli <USER>@<SERVER_IP>:/tmp/sl-cli-new
ssh <USER>@<SERVER_IP> "sudo cp /tmp/sl-cli-new /home/<USER>/sl-cli && sudo chown <USER> /home/<USER>/sl-cli"
```

### 2. Start Daemon

```bash
ssh <USER>@<SERVER_IP> "sudo -u <USER> nohup /home/<USER>/sl-cli backend start \
  --daemon --port 3100 --no-systemd --auth-token <TOKEN> > /dev/null 2>&1 &"
```

### 3. Sync a Site

```bash
# Auto-creates site/version on remote, no manual scp needed
sl-cli site sync <slug> --host <SERVER_IP> --user <USER> --key <KEY>
```

Or via HTTP API (after daemon is running):
```bash
sl-cli site sync <slug> --target dk2
```

### 4. Hotify Config Sync (required for DNS/Traefik)

The daemon user must have the same hotify config as the infra user:

```bash
scp ~/.hotify/config.json <USER>@<SERVER_IP>:/tmp/
ssh <USER>@<SERVER_IP> "mkdir -p ~/.hotify && cp /tmp/hotify-config.json ~/.hotify/config.json"
```

### 5. Fix Traefik Config Ownership

```bash
ssh <USER>@<SERVER_IP> "sudo chown <USER> /etc/traefik/dynamic.yml /etc/traefik/traefik.yml /etc/traefik/cloudflare.env"
ssh <USER>@<SERVER_IP> "echo '<USER> ALL=(ALL) NOPASSWD: /usr/bin/systemctl restart traefik' | sudo tee /etc/sudoers.d/<USER>-traefik"
```

### 6. Configure DNS & Traefik via Remote API

```bash
sl-cli site dns setup <site> --domain <domain> --ip <IP> --traefik --target dk2
```

Or directly on the server:
```bash
ssh <USER>@<SERVER_IP> "hotify-cli setup-dns --id <app> --ip <IP> --local"
ssh <USER>@<SERVER_IP> "hotify-cli setup-traefik --id <app> --challenge-type dns --local"
```

## Troubleshooting

| Issue | Check | Fix |
|-------|-------|-----|
| 404 on domain | Traefik routing | Add router in `/etc/traefik/dynamic.yml` with `addPrefix: /<slug>` |
| SSL cert fails | CF token on remote | Sync `~/.hotify/config.json` |
| Traefik config denied | File ownership | `sudo chown <USER> /etc/traefik/*` |
| Traefik restart denied | Sudoers | Add NOPASSWD sudoers rule for `systemctl restart traefik` |
| Daemon not responding | Binary out of date | Rebuild + scp + restart |
| Port 443 blocked | Tailscale Funnel | `tailscale funnel reset` |
| Port 443 blocked | iptables redirect | `iptables -t nat -D PREROUTING -p tcp --dport 443 -j REDIRECT --to-port 8443` |
| Domain duplication | hotify base domain | Use subdomain only in `--domain` |

## Key Gotchas

- **hotify config**: Daemon user needs the same `~/.hotify/config.json` as the infra user
- **Traefik permissions**: Daemon user needs write access to `/etc/traefik/*.yml` + passwordless sudo for `systemctl restart traefik`
- **No daemon restart after sync**: Files are served from disk on each request — sync changes are live immediately
- **Dynamic blocks work**: `{{>include "path"}}` is fully implemented, processed at serve time
- **Admin user**: Daemon runs as a specific user (e.g., `admin` on dk2). All permissions must be set for that user.

## References

- hotify-integration skill for hotify-cli details
- superlandings-go-assets skill for asset/CLI conventions
- sync-mechanism skill for sync internals
