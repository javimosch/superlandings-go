---
name: superlandings-go-assets
description: Asset management, CLI output conventions, and template helpers for SuperLandings Go
---

# Asset Management & CLI Conventions

## Asset Storage

Assets (images, CSS, JS) are stored in `~/.superlandings/sites/{slug}/assets/` — shared across all versions. No duplication on version rollback.

## CLI Commands

```bash
# Upload (create/replace)
sl-cli site upload <site> "<path>" --file <local-file>
sl-cli site upload my-site "logo.png" --file ./logo.png
sl-cli site upload my-site "css/style.css" --file ./dist/style.css

# List
sl-cli site assets list <site>

# Remove
sl-cli site assets remove <site> "<path>"

# Remote (any command + --target <host:port>)
sl-cli site upload my-site "img.png" --file ./img.png --target dk2
```

## Template Helper: `{{asset "filename"}}`

Recursively searches the shared assets directory by filename and returns the first match as a URL path.

```html
<img src="{{asset "logo.png"}}" alt="Logo">
<link rel="stylesheet" href="{{asset "style.css"}}">
```

If `assets/logo.png` and `assets/nav/logo.png` both exist, resolves to `/slug/logo.png` (first found).

## Output Convention

All commands output JSON by default:

| Type | Structure |
|------|-----------|
| Success | `{"version":"1.0","success":true,"message":"...","id":"...","slug":"..."}` |
| Data | `{"version":"1.0","sites":[...],"assets":[...]}` |
| Error | `{"version":"1.0","success":false,"error":{"code":81,"type":"invalid_input","message":"...","recoverable":false}}` |

Errors go to stderr with semantic exit codes:

| Code | Meaning | Example |
|------|---------|--------|
| 81 | Missing flag | `--name is required` |
| 90 | Not found | `site not found` |
| 92 | Conflict | `slug already exists` |
| 101 | External failure | `hotify-cli error` |

## Domain-Aware Serving

The daemon serves content from domains without Traefik path rewriting:

```
https://test-site.intrane.fr/blue.png
  → Traefik forwards to :3100/blue.png
  → daemon sees Host: test-site.intrane.fr
  → looks up site_domains → resolves slug "test-site"
  → serves sites/test-site/assets/blue.png ✅
```

## Remote Execution Flow

Local `sl-cli --target <host:port>` → HTTP POST to remote daemon API → remote process action → JSON response.

Asset upload via `--target`: reads file locally → base64 encodes → POSTs to `/api/sites/{slug}/upload` → remote daemon decodes and writes to disk.

## Deploying Updated Binary

```bash
scp sl-cli <user>@<host>:~/sl-cli-new
ssh <user>@<host> "sudo -u <user> pkill -f 'sl-cli backend'; \
  sudo cp ~/sl-cli-new /home/<user>/sl-cli; \
  sudo -u <user> nohup /home/<user>/sl-cli backend start --daemon --port 3100 --no-systemd > /dev/null 2>&1 &"
```

## Gotchas

- **hotify config**: The daemon user needs the same `~/.hotify/config.json` as the infra user or DNS/Traefik setup fails.
- **Traefik permissions**: `/etc/traefik/*.yml` must be writable by the daemon user + passwordless sudo for `systemctl restart traefik`.
- **Template functions before Parse**: Go's `html/template` requires `.Funcs()` registered before `.Parse()`.
