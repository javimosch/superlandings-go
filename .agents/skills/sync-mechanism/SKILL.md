---
name: sync-mechanism
description: SSH-based site sync and proxy mechanism for SuperLandings Go
---

# Site Sync & Proxy Mechanism

## Overview

SuperLandings Go includes SSH-based site synchronization and proxy configuration for deploying sites to remote servers.

## Sync Commands

### Export Site Metadata

```bash
sl-cli site export <site> --output /tmp/export.json
```

Exports site metadata (sites, versions, files) to JSON for migration or backup.

### Import Site Metadata

```bash
sl-cli site import --input /tmp/import.json
```

Imports site metadata from JSON file. Creates site and versions if they don't exist.

### Sync to Remote Server

```bash
sl-cli site sync <site> --host <host> --user <user> [--port 22] [--key <ssh-key-path>]
```

Synchronizes a site to a remote server:
1. Exports site metadata to JSON
2. Rsyncs site files to remote
3. Copies JSON to remote
4. Imports metadata on remote
5. Restarts daemon on remote

**Example:**
```bash
sl-cli site sync slv2 --host 92.113.145.16 --user root --key ~/.ssh/id_rsa_srv
```

## Proxy Commands

### Setup Hotify Proxy

```bash
sl-cli site proxy <site> --domain <domain> --internal-url <url>
```

Configures hotify-cli reverse proxy for a site:
1. Calls hotify-cli setup with domain and backend-url
2. Calls hotify-cli setup-traefik for ACME certificates

**Example:**
```bash
sl-cli site proxy slv2 --domain slv2.intrane.fr --internal-url http://127.0.0.1:3100
```

**Limitation:** Currently requires manual Traefik configuration for path prefix middleware. See hotify-integration skill.

## Implementation Details

### Sync Service (internal/services/sync.go)

**SyncTarget struct:**
```go
type SyncTarget struct {
    Host string
    User string
    Port int
    Key  string // SSH key path
}
```

**Sync workflow:**
1. Export site metadata to temporary JSON file
2. Rsync site directory using SSH key
3. SCP export file to remote
4. SSH to remote and run import command
5. SSH to remote and restart daemon

**SSH key support:**
- Uses `-i <key> -o IdentitiesOnly=yes` for rsync, scp, ssh
- Falls back to default SSH config if no key specified

### CLI Commands (internal/cli/site_sync.go)

**site export:**
- Flags: `--output` (default: /tmp/site-export.json)
- Exports site metadata to JSON file

**site import:**
- Flags: `--input` (required)
- Imports site metadata from JSON file

**site sync:**
- Flags: `--host` (required), `--user` (default: root), `--port` (default: 22), `--key`
- Syncs site to remote target

**site proxy:**
- Flags: `--domain` (required), `--internal-url` (default: http://127.0.0.1:3099)
- Configures hotify-cli reverse proxy

## Current Limitations

1. **Sync command rsync issues**
   - May fail with SSH key authentication
   - Manual sync (scp) is more reliable currently

2. **Proxy command Traefik issues**
   - hotify-cli setup-traefik needs sudo
   - Doesn't generate router/service config for backend-url
   - Requires manual Traefik configuration

3. **Import requires site to exist**
   - Cannot create site during import
   - Must create site first, then import versions

## Manual Sync Workaround

When automated sync fails:

```bash
# 1. Create site directory on remote
ssh -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv dk2 "mkdir -p /home/dk2/.superlandings/sites/slv2/v1"

# 2. Copy site files
scp -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv -r ~/.superlandings/sites/slv2/* dk2:/home/dk2/.superlandings/sites/slv2/

# 3. Create version on remote
ssh -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv dk2 "sl-cli site version create slv2 --version v1 --comment 'Synced from local'"

# 4. Restart daemon
ssh -o IdentitiesOnly=yes -i ~/.ssh/id_rsa_srv dk2 "pkill -f 'sl-cli backend' && sl-cli backend start --daemon --port 3100"
```

## File Structure

```
internal/
├── services/
│   └── sync.go          # Sync service implementation
├── cli/
│   └── site_sync.go     # CLI commands (export, import, sync, proxy)
└── db/
    └── site_version_repository.go  # Version repository (split from repository.go)
```

## LOC Limits

All files kept under 400 LOC per AGENTS.md rules:
- sync.go: ~240 LOC
- site_sync.go: ~175 LOC
- site_version_repository.go: ~109 LOC

## References

- hotify-integration skill for hotify-cli details
- dk2-deployment skill for full deployment workflow
- AGENTS.md for coding guidelines