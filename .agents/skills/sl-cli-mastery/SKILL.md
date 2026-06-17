---
name: sl-cli-mastery
description: Create sites with templates, assets, blog, admin panel — full mastery guide
---

# sl-cli Mastery

## Quick Site (60s)

```bash
sl-cli site create --name "My Site" --slug "my-site"
sl-cli site version create my-site --version "v1"
sl-cli site upload my-site "logo.png" --file ./logo.png
sl-cli site write my-site v1 "index.html" --content '{{>include "nav.html"}}<h1>{{.title}}</h1>{{asset "logo.png"}}'
echo '{"title":"My Site"}' > ~/.superlandings/sites/my-site/v1/index.html.data.json
sl-cli backend start --daemon --port 3099 --no-systemd
```

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

### Admin Panel

**Enable per site via `admin-schema.json`** (at site level, not in version dir):

```bash
sl-cli site admin configure <site> --auto-detect  # auto-generate schema
sl-cli site admin create <site>                    # generate token
sl-cli user create --email admin --password pass   # create user
sl-cli user grant <site> admin                     # grant access
```

### Auth Modes
- `"auth": "none"` — `/admin/{slug}/{token}`, tokens never expire
- `"auth": "password"` — `/admin/{slug}` shows login form, JWT sessions, logout

### Editor Types
- **Raw HTML**: `"type": "form"`, `source: "*.html"` → CodeMirror (configurable `layout.editorHeight`)
- **Form fields**: `"type": "form"`, `source: "*.data.json"` → text/textarea for non-technical users
- **Blog**: `"type": "markdown"`, `source: "blog"` → EasyMDE editor with metadata, drafts, delete

## Blog Module

```bash
mkdir -p ~/.superlandings/sites/{slug}/v1/blog
# Write post.md + post.md.data.json {"title":"...","published":true}
# Add {{>include "blog-preview.html"}} to your index.html
# Create layout.html for blog post styling (nav, footer, CSS)
```

Posts auto-discovered from `blog/*.md`. Drafts hidden (`published:false`). Metadata: `title`, `author`, `date`, `reading_time`.

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
