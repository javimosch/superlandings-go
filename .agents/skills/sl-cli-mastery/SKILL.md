---
name: sl-cli-mastery
description: Create a complete landing page with assets, templates, and includes in under 60 seconds
---

# sl-cli Mastery — Full Landing in <60s

## 60-Second Landing Recipe

```bash
# 1. Create site + version (5s)
sl-cli site create --name "My Site" --slug "my-site"
sl-cli site version create my-site --version "v1"

# 2. Upload assets (10s)
sl-cli site upload my-site "logo.png" --file ./logo.png
sl-cli site upload my-site "style.css" --file ./style.css

# 3. Write templates (30s)
sl-cli site write my-site v1 "nav.html" --content '<nav>{{asset "logo.png"}}<a href="/">Home</a></nav>'
sl-cli site write my-site v1 "index.html" --content '{{>include "nav.html"}}<h1>{{.title}}</h1>{{asset "banner.jpg"}}'
sl-cli site write my-site v1 "footer.html" --content '<footer>© {{.year}}</footer>'

# 4. Write data file (5s)
echo '{"title":"My Site","year":"2026"}' > ~/.superlandings/sites/my-site/v1/index.html.data.json

# 5. Serve (5s)
sl-cli backend start --daemon --port 3099 --no-systemd
curl http://localhost:3099/my-site/
```

Total: **~55 seconds** from zero to live site.

## Template System

### Processing Order
1. **Layout** `{{>layout "layout.html"}}` — wraps content, replaces `{{.content}}`
2. **Includes** `{{>include "nav.html"}}` — inlines file, recursively processed
3. **Go templates** — `{{.var}}`, `{{if}}`, `{{range}}`, `{{asset "file"}}`

### Asset Helper `{{asset "filename"}}`
- Searches the entire `assets/` directory tree by filename
- Returns the first match as a URL path
- `{{asset "logo.png"}}` → `/my-site/logo.png` or `/my-site/img/logo.png` (whichever found first)
- Works with any nesting depth

### Data File `index.html.data.json`
```json
{
  "title": "My Site",
  "items": [{"name": "A"}, {"name": "B"}]
}
```
```html
<h1>{{.title}}</h1>
{{range .items}}<p>{{.name}}</p>{{end}}
```

### Common Patterns

```html
<!-- Auto-discovered navigation -->
{{range .nav_pages}}<a href="/my-site/{{.slug}}">{{.title}}</a>{{end}}

<!-- Auto-discovered blog posts -->
{{range .blog_posts}}<h2>{{.title}}</h2>{{end}}

<!-- Auto-injected root path -->
<a href="{{.root}}/page">link</a>

<!-- Auto-injected brand name -->
<div class="brand">{{.brand}}</div>
```

## Asset CRUD

```bash
sl-cli site upload <site> "<path>" --file <local>    # create / replace
sl-cli site assets list <site>                        # list
sl-cli site assets remove <site> "<path>"             # delete
```

All operations support `--target <host:port>` for remote execution.

## Remote Execution

```bash
# Add target once
sl-cli targets add --name dk2 --host <IP> --port 3100 --token <TOKEN>

# Then any command works remotely
sl-cli site list --target dk2
sl-cli site upload my-site "img.png" --file ./img.png --target dk2
sl-cli site assets list my-site --target dk2
```

## Serving Modes

| URL | How it resolves |
|-----|----------------|
| `localhost:3099/slug/path` | Path-based: slug from URL |
| `domain.com/path` | Host header lookup in `site_domains` table |
| Direct port | `http://<IP>:3100/slug/path` |

## Key Pitfalls

- **Template functions before Parse**: Go's `html/template` requires `.Funcs()` before `.Parse()`. The `{{asset}}` helper is registered this way — if you add custom functions, do the same.
- **No `{{range}}` as example text**: Go's template engine will interpret it as a real action. Use plain text descriptions instead.
- **Assets shared across versions**: Upload once, all versions see it. No duplication.
- **Data file name**: Must be `filename.html.data.json` (e.g., `index.html.data.json`).
- **Includes are recursive**: Files included via `{{>include}}` can include other files.
- **All output is JSON**: Commands return `{"version":"1.0","success":true,...}`. Errors on stderr with semantic exit codes (81=missing flag, 90=not found, 101=external error).

## Full Example (copy-paste ready)

```bash
# Create
sl-cli site create --name "Demo" --slug "demo"
sl-cli site version create demo --version "v1"

# Assets
echo 'body{font-family:sans-serif}' > /tmp/style.css
python3 -c "open('/tmp/logo.png','wb').write(open('/dev/urandom','rb').read(100))"
sl-cli site upload demo "style.css" --file /tmp/style.css
sl-cli site upload demo "logo.png" --file /tmp/logo.png

# Templates
sl-cli site write demo v1 "nav.html" --content '<nav><img src="{{asset \"logo.png\"}}" height="32"> <a href="/">Home</a></nav>'
sl-cli site write demo v1 "index.html" --content '<!DOCTYPE html><html><body>{{>include "nav.html"}}<h1>{{.title}}</h1><p>{{.desc}}</p></body></html>'

# Data
echo '{"title":"Demo","desc":"Created in under 60s"}' > ~/.superlandings/sites/demo/v1/index.html.data.json

# Serve
sl-cli backend start --daemon --port 3099 --no-systemd
curl http://localhost:3099/demo/
```

## References
- `superlandings-go-assets` skill for asset details
- `hotify-integration` skill for DNS/Traefik
- `sync-mechanism` skill for SSH sync
- `AGENTS.md` for project structure and limits
