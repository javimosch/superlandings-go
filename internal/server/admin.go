package server

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/javimosch/superlandings-go/internal/services"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("superlandings-secret-key") // TODO: Move to config

// handleAdmin serves the admin login page or editor UI
func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/admin/")
	parts := strings.Split(path, "/")

	if len(parts) < 1 || parts[0] == "" {
		http.Error(w, "Invalid admin URL", http.StatusBadRequest)
		return
	}

	siteSlug := parts[0]

	// Get site
	siteRepo := db.NewSiteRepository()
	site, err := siteRepo.GetBySlug(siteSlug)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Read schema to check if auth is required
	schemaPath := filepath.Join(s.cfg.SitesDir, site.Slug, "admin-schema.json")
	authRequired := false
	if data, err := os.ReadFile(schemaPath); err == nil {
		var schema map[string]interface{}
		if json.Unmarshal(data, &schema) == nil {
			if a, ok := schema["auth"].(string); ok && a == "password" {
				authRequired = true
			}
		}
	}

	// Auth sites: let login handler process POST requests
	if authRequired && r.Method == "POST" {
		s.handleAdminLogin(w, r, site)
		return
	}

	// Auth sites: check for existing JWT session first
	if authRequired && r.Method == "GET" {
		sessionCookie, err := r.Cookie("sl_admin_session")
		if err == nil {
			claims, err := validateJWT(sessionCookie.Value)
			if err == nil && claims.SiteID == site.ID {
				s.handleAdminEditor(w, r, site)
				return
			}
		}
	}

	// Auth sites: /admin/slug is enough (no token needed, login form is the gate)
	if authRequired && len(parts) < 2 {
		s.handleAdminLogin(w, r, site)
		return
	}

	// No-auth sites need a token
	if !authRequired && len(parts) < 2 {
		http.Error(w, "This site requires an admin token. Use sl-cli site admin create "+siteSlug, http.StatusBadRequest)
		return
	}

	// Auth sites with token: token IS the authentication
	if authRequired {
		token := parts[1]
		adminRepo := db.NewSiteAdminRepository()
		adminToken, err := adminRepo.GetTokenByValue(token)
		if err == nil && adminToken.IsActive {
			s.handleAdminEditor(w, r, site)
			return
		}
		// Invalid token, fall through to login
		s.handleAdminLogin(w, r, site)
		return
	}

	// No-auth sites with token: verify and go directly to editor
	token := parts[1]
	adminRepo := db.NewSiteAdminRepository()
	adminToken, err := adminRepo.GetTokenByValue(token)
	if err != nil || !adminToken.IsActive {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}
	// Token is valid — go directly to editor
	s.handleAdminEditor(w, r, site)
}

// handleAdminLogout clears the JWT session cookie and redirects to login
func (s *Server) handleAdminLogout(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Query().Get("slug")
	http.SetCookie(w, &http.Cookie{
		Name:     "sl_admin_session",
		Value:    "",
		Path:     fmt.Sprintf("/admin/%s", slug),
		MaxAge:   -1,
		HttpOnly: true,
	})
	http.Redirect(w, r, "/admin/"+slug, http.StatusSeeOther)
}

// handleAdminLogin serves the login form
func (s *Server) handleAdminLogin(w http.ResponseWriter, r *http.Request, site *db.Site) {
	if r.Method == "GET" {
		html := `<!DOCTYPE html>
<html>
<head>
	<title>Login - ` + site.Name + `</title>
	<style>
		body { font-family: system-ui, sans-serif; display: flex; justify-content: center; align-items: center; min-height: 100vh; background: #f5f5f5; }
		.login-box { background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); width: 100%; max-width: 400px; }
		h2 { margin-top: 0; color: #333; }
		input { width: 100%; padding: 0.75rem; margin: 0.5rem 0; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
		button { width: 100%; padding: 0.75rem; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; margin-top: 1rem; }
		button:hover { background: #0056b3; }
		.error { color: #dc3545; margin-top: 1rem; }
	</style>
</head>
<body>
	<div class="login-box">
		<h2>Login to ` + site.Name + `</h2>
		<form method="POST">
			<input type="email" name="email" placeholder="Email" required>
			<input type="password" name="password" placeholder="Password" required>
			<button type="submit">Login</button>
		</form>
	</div>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
		return
	}

	if r.Method == "POST" {
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Verify credentials
		userRepo := db.NewUserRepository()
		valid, err := userRepo.VerifyPassword(email, password)
		if err != nil || !valid {
			html := `<!DOCTYPE html>
<html>
<head>
	<title>Login - ` + site.Name + `</title>
	<style>
		body { font-family: system-ui, sans-serif; display: flex; justify-content: center; align-items: center; min-height: 100vh; background: #f5f5f5; }
		.login-box { background: white; padding: 2rem; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); width: 100%; max-width: 400px; }
		h2 { margin-top: 0; color: #333; }
		input { width: 100%; padding: 0.75rem; margin: 0.5rem 0; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
		button { width: 100%; padding: 0.75rem; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; margin-top: 1rem; }
		button:hover { background: #0056b3; }
		.error { color: #dc3545; margin-top: 1rem; }
	</style>
</head>
<body>
	<div class="login-box">
		<h2>Login to ` + site.Name + `</h2>
		<form method="POST">
			<input type="email" name="email" placeholder="Email" required>
			<input type="password" name="password" placeholder="Password" required>
			<button type="submit">Login</button>
			<div class="error">Invalid email or password</div>
		</form>
	</div>
</body>
</html>`
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(html))
			return
		}

		// Get user
		user, err := userRepo.GetByEmail(email)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Check if user has access to this site
		siteUsers, err := userRepo.GetSiteUsers(site.ID)
		if err != nil || len(siteUsers) == 0 {
			http.Error(w, "You don't have access to this site", http.StatusForbidden)
			return
		}

		hasAccess := false
		for _, su := range siteUsers {
			if su.UserID == user.ID {
				hasAccess = true
				break
			}
		}

		if !hasAccess {
			http.Error(w, "You don't have access to this site", http.StatusForbidden)
			return
		}

		// Create session token (JWT)
		sessionToken, err := createJWT(user.ID, site.ID, 24*time.Hour)
		if err != nil {
			http.Error(w, "Failed to create session", http.StatusInternalServerError)
			return
		}

		// Set cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "sl_admin_session",
			Value:    sessionToken,
			Path:     "/admin/" + site.Slug,
			MaxAge:   86400, // 24 hours
			HttpOnly: true,
			Secure:   false, // TODO: Set to true in production with HTTPS
			SameSite: http.SameSiteStrictMode,
		})

		// Redirect to same URL (cookie is set, next GET shows editor)
		http.Redirect(w, r, r.URL.Path, http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleAdminEditor serves the editor UI
func (s *Server) handleAdminEditor(w http.ResponseWriter, r *http.Request, site *db.Site) {
	schemaPath := filepath.Join(s.cfg.SitesDir, site.Slug, "admin-schema.json")
	schemaJSON := `{"sections":[]}`
	if data, err := os.ReadFile(schemaPath); err == nil {
		schemaJSON = string(data)
	}

	escapedSchema := strings.ReplaceAll(schemaJSON, "\\", "\\\\")
	escapedSchema = strings.ReplaceAll(escapedSchema, "'", "\\'")

	html := `<!DOCTYPE html>
<html>
<head>
	<title>` + site.Name + ` &mdash; Editor</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/easymde@2.18.0/dist/easymde.min.css">
	<script src="https://cdn.jsdelivr.net/npm/easymde@2.18.0/dist/easymde.min.js"></script>
	<style>
		:root{--primary:#2563eb;--bg:#f8fafc;--card:#fff;--text:#1e293b;--muted:#94a3b8;--border:#e2e8f0}
		*{margin:0;padding:0;box-sizing:border-box}
		body{font-family:system-ui,sans-serif;background:var(--bg);color:var(--text);line-height:1.6}
		.hdr{background:var(--card);border-bottom:1px solid var(--border);padding:.75rem 1.5rem;display:flex;align-items:center;justify-content:space-between;position:sticky;top:0;z-index:10}
		.hdr h1{font-size:1.1rem;font-weight:600}.hdr .site{color:var(--muted);font-weight:400}
		.hdr a{color:var(--primary);text-decoration:none;font-size:.9rem}
		.wrap{display:flex;height:calc(100vh - 56px)}
		.sidebar{width:260px;background:var(--card);border-right:1px solid var(--border);padding:1rem;overflow-y:auto;flex-shrink:0}
		.sidebar h2{font-size:.75rem;text-transform:uppercase;letter-spacing:.05em;color:var(--muted);margin:1rem 0 .5rem}
		.section-btn{display:block;width:100%;text-align:left;padding:.5rem .75rem;border:none;background:transparent;border-radius:6px;cursor:pointer;font-size:.875rem;color:var(--text);font-weight:500;transition:background .15s}
		.section-btn:hover,.section-btn.active{background:#eff6ff;color:var(--primary)}
		.main{flex:1;display:flex;flex-direction:column;overflow:hidden}
		.section-panel{flex:1;display:none;flex-direction:column;overflow:hidden}
		.section-panel.active{display:flex}
		/* Empty */
		.empty{flex:1;display:flex;flex-direction:column;align-items:center;justify-content:center;color:var(--muted);padding:2rem;text-align:center}
		.empty h2{font-size:1.25rem;color:var(--text);margin-bottom:.5rem}
		.empty p{font-size:.9rem;margin-bottom:1.5rem;max-width:400px}
		/* Markdown editor */
		.editor-toolbar{display:flex;align-items:center;gap:.75rem;padding:.75rem 1.5rem;border-bottom:1px solid var(--border);background:var(--card)}
		.editor-toolbar input{flex:1;border:none;font-size:1.1rem;font-weight:600;outline:none;background:transparent;color:var(--text)}
		.editor-toolbar input.meta{font-size:.85rem;font-weight:400;color:var(--muted)}
		.EasyMDEContainer{border:none!important;border-radius:0!important;flex:1;display:flex;flex-direction:column}
		.EasyMDEContainer .editor-toolbar{border:none!important;border-bottom:1px solid var(--border)!important}
		.EasyMDEContainer .CodeMirror{flex:1!important;border:none!important;border-radius:0!important;font-size:.95rem!important}
		/* Form editor */
		.form-grid{display:grid;gap:1rem;padding:1.5rem;max-width:600px}
		.form-grid label{font-size:.85rem;font-weight:500;color:var(--muted);display:block;margin-bottom:.25rem}
		.form-grid input,.form-grid textarea{width:100%;padding:.6rem .75rem;border:1px solid var(--border);border-radius:6px;font-size:.9rem;outline:none;font-family:inherit}
		.form-grid textarea{min-height:100px;resize:vertical}
		.form-grid input:focus,.form-grid textarea:focus{border-color:var(--primary);box-shadow:0 0 0 3px rgba(37,99,235,.1)}
		/* Post list */
		.post-list{list-style:none}
		.post-list li{padding:.5rem .75rem;border-radius:6px;cursor:pointer;font-size:.875rem;display:flex;align-items:center;justify-content:space-between;transition:background .15s}
		.post-list li:hover{background:#f1f5f9}
		.post-list li .tt{font-weight:500}.post-list li .tag{font-size:.65rem;background:#e0f2fe;color:#0284c7;padding:.15rem .35rem;border-radius:3px}
		/* Buttons */
		.btn{display:inline-flex;align-items:center;gap:.4rem;padding:.5rem 1rem;border-radius:6px;font-size:.875rem;font-weight:500;border:none;cursor:pointer;transition:all .15s}
		.btn-primary{background:var(--primary)!important;color:#fff!important}.btn-primary:hover{background:#1d4ed8!important}
		.btn-success{background:#059669!important;color:#fff!important}.btn-success:hover{background:#047857!important}
		.btn-sm{padding:.35rem .75rem;font-size:.8rem}
		.toast{position:fixed;bottom:1.5rem;right:1.5rem;background:#065f46;color:#fff;padding:.75rem 1.25rem;border-radius:8px;font-size:.875rem;box-shadow:0 4px 12px rgba(0,0,0,.15);opacity:0;transform:translateY(10px);transition:all .3s;z-index:200}
		.toast.show{opacity:1;transform:translateY(0)}
		/* Schema warning */
		.schema-warn{text-align:center;padding:2rem;color:var(--muted)}
		.schema-warn code{display:block;margin:1rem;background:#f1f5f9;padding:.5rem;border-radius:4px;font-size:.85rem}
	</style>
</head>
<body>
<div class="hdr">
	<h1><span class="site">` + site.Name + `</span> Editor</h1>
	<div style="display:flex;align-items:center;gap:.75rem">
		<span id="auth-state" style="font-size:.8rem;color:var(--muted)"></span>
		<a href="javascript:logout()" style="color:var(--muted);text-decoration:none;font-size:.85rem">Logout</a>
		<a href="/` + site.Slug + `" target="_blank">View site &rarr;</a>
	</div>
</div>
<div class="wrap" id="app">
	<div class="sidebar" id="sidebar">
		<h2>Sections</h2>
		<div id="section-nav"></div>
	</div>
	<div class="main">
		<div id="section-content"></div>
	</div>
</div>
<div class="toast" id="toast">Saved!</div>

<script id="admin-schema" type="application/json">` + schemaJSON + `</script>
<script>
const slug='` + site.Slug + `';
const schema=JSON.parse(document.getElementById('admin-schema').textContent);
const sects=schema.sections||[];
let easyMDE=null,currentPost=null;

function toast(m){const t=document.getElementById('toast');t.textContent=m;t.classList.add('show');setTimeout(()=>t.classList.remove('show'),2000)}

function buildUI(){
	const nav=document.getElementById('section-nav'),content=document.getElementById('section-content');
	if(!sects.length){nav.innerHTML='';content.innerHTML='<div class="schema-warn"><h2>No admin schema configured</h2><p>Run <code>sl-cli admin configure '+slug+' --auto-detect</code> to generate one.</p></div>';return;}

	nav.innerHTML=sects.map((s,i)=>'<button class="section-btn'+(i===0?' active':'')+'" onclick="showSection('+i+')">'+s.title+'</button>').join('');
	content.innerHTML=sects.map((s,i)=>'<div class="section-panel'+(i===0?' active':'')+'" id="panel-'+i+'"></div>').join('');
	
	sects.forEach((s,i)=>{
		var p=document.getElementById('panel-'+i);
		if(!p)return;
		if(s.type==='markdown') renderMarkdown(p,s);
		else if(s.type==='form') renderForm(p,s);
	});
}

function showSection(i){
	document.querySelectorAll('.section-btn').forEach(b=>b.classList.remove('active'));
	document.querySelectorAll('.section-panel').forEach(p=>p.classList.remove('active'));
	document.querySelectorAll('.section-btn')[i].classList.add('active');
	document.getElementById('panel-'+i).classList.add('active');
}

/* === MARKDOWN SECTION === */
function renderMarkdown(panel,sec){
	panel.innerHTML='<div style="display:flex;flex:1;overflow:hidden"><div id="blog-editor-area" style="display:none;flex:1;flex-direction:column;overflow:hidden"><div style="padding:.75rem 1.5rem;background:var(--card);border-bottom:1px solid var(--border)"><input type="text" id="post-title" placeholder="Post title..." style="width:100%;border:none;font-size:1.1rem;font-weight:600;outline:none;background:transparent;color:var(--text);margin-bottom:.5rem"><div style="display:flex;gap:.75rem;align-items:flex-end;flex-wrap:wrap"><div style="display:flex;flex-direction:column;gap:.15rem"><small style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.05em">Author</small><input id="post-author" class="meta" placeholder="Name" style="width:150px;border:1px solid var(--border);border-radius:6px;padding:.35rem .5rem;font-size:.85rem;outline:none;background:var(--card);color:var(--text)"></div><div style="display:flex;flex-direction:column;gap:.15rem"><small style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.05em">Date</small><input id="post-date" class="meta" placeholder="2026-01-01" style="width:130px;border:1px solid var(--border);border-radius:6px;padding:.35rem .5rem;font-size:.85rem;outline:none;background:var(--card);color:var(--text)"></div><div style="display:flex;flex-direction:column;gap:.15rem"><small style="color:var(--muted);font-size:.7rem;text-transform:uppercase;letter-spacing:.05em">Read</small><input id="post-time" class="meta" placeholder="4 min" style="width:80px;border:1px solid var(--border);border-radius:6px;padding:.35rem .5rem;font-size:.85rem;outline:none;background:var(--card);color:var(--text)"></div><label style="display:flex;align-items:center;gap:.3rem;font-size:.8rem;white-space:nowrap"><input type="checkbox" id="post-published" checked> Published</label><button class="btn btn-primary btn-sm" onclick="savePost()">Publish</button><button class="btn btn-outline btn-sm" onclick="deletePost()" style="color:#dc2626">Delete</button></div></div><textarea id="markdown-editor"></textarea></div><div class="sidebar" style="border-left:1px solid var(--border);border-right:none"><h2>Posts</h2><ul class="post-list" id="post-list"></ul><button class="btn btn-success btn-sm" style="width:100%;margin-top:.5rem" onclick="newPost()">+ New Post</button></div></div>';
	loadPosts();
}

function loadPosts(){
	var list=document.getElementById('post-list');if(!list)return;
	fetch('/api/sites/'+slug+'/files?path=blog').then(function(r){return r.json()}).then(function(d){
		list.innerHTML='';
		(d.files||[]).forEach(function(f){
			if(!f.name.endsWith('.md'))return;
			var n=f.name.replace(/\.md$/,'').replace(/-/g,' ');var lbl=n.charAt(0).toUpperCase()+n.slice(1);
			var li=document.createElement('li');
			li.innerHTML='<span class="tt">'+lbl+'</span>';
			li.onclick=function(){editPost(f.path);};
			list.appendChild(li);
		});
	});
}

function editPost(path){currentPost=path;
	var metaPath=path+'.data.json';
	// Load metadata
	fetch('/api/sites/'+slug+'/files/'+metaPath).then(function(r){return r.json()}).then(function(d){
		try{
			var meta=JSON.parse(d.content);
			document.getElementById('post-title').value=meta.title||'';
			document.getElementById('post-author').value=meta.author||'';
			document.getElementById('post-date').value=meta.date||'';
			document.getElementById('post-time').value=meta.reading_time||'';
			document.getElementById('post-published').checked=meta.published!==false;
		}catch(e){}
	}).catch(function(){});
	// Load markdown content
	fetch('/api/sites/'+slug+'/files/'+path).then(function(r){return r.json()}).then(function(d){
		document.getElementById('blog-editor-area').style.display='flex';
		var lines=d.content.split('\n');var title='',body=d.content;
		for(var i=0;i<lines.length;i++){if(lines[i].startsWith('# ')){title=lines[i].replace(/^# /,'').trim();body=lines.slice(i+1).join('\n').trim();break;}}
		if(!document.getElementById('post-title').value)document.getElementById('post-title').value=title||'Untitled';
		if(easyMDE)easyMDE.value(body);else{
			easyMDE=new EasyMDE({element:document.getElementById('markdown-editor'),spellChecker:false,autofocus:true,placeholder:'Write your post...',status:false,toolbar:['bold','italic','heading','|','quote','unordered-list','ordered-list','|','link','image','|','preview','side-by-side','fullscreen']});
			easyMDE.value(body);
		}
	});
}

function savePost(){
	var title=document.getElementById('post-title').value.trim();
	var body=easyMDE?easyMDE.value().trim():'';
	if(!title&&!body){toast('Nothing to save');return;}
	var content=(title?'# '+title+'\n\n':'')+body;
	var fp=currentPost;
	if(!fp){var sn=title.toLowerCase().replace(/[^a-z0-9]+/g,'-').replace(/^-|-$/g,'')||'untitled';fp='blog/'+sn+'.md';currentPost=fp;}

	var btn=document.getElementById('post-title');btn.disabled=true;
	fetch('/api/sites/'+slug+'/write',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({file:fp,content:content})})
	.then(function(r){return r.json()}).then(function(){
		// Save metadata
		var meta={title:title,author:document.getElementById('post-author').value.trim(),date:document.getElementById('post-date').value.trim(),reading_time:document.getElementById('post-time').value.trim(),published:document.getElementById('post-published').checked};
		fetch('/api/sites/'+slug+'/write',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({file:fp+'.data.json',content:JSON.stringify(meta)})})
		.then(function(r){return r.json()}).then(function(){toast('Published!');btn.disabled=false;loadPosts();});
	});
}

function deletePost(){
	if(!currentPost){toast('No post selected');return;}
	if(!confirm('Delete this post permanently?'))return;
	var meta=currentPost+'.data.json';
	fetch('/api/sites/'+slug+'/files/'+currentPost,{method:'DELETE'})
	.then(function(){return fetch('/api/sites/'+slug+'/files/'+meta,{method:'DELETE'})})
	.then(function(){document.getElementById('blog-editor-area').style.display='none';currentPost=null;loadPosts();toast('Deleted');});
}

function newPost(){if(easyMDE)easyMDE.value('');document.getElementById('post-title').value='';document.getElementById('post-author').value='';document.getElementById('post-date').value=new Date().toISOString().slice(0,10);document.getElementById('post-time').value='3 min';document.getElementById('post-published').checked=true;document.getElementById('blog-editor-area').style.display='flex';currentPost=null;setTimeout(function(){document.getElementById('post-title').focus()},100);}

// === FORM SECTION ===
function renderForm(panel,sec){
	if(!panel)return;
	const fields=sec.fields||[];
	const source=sec.source||'index.html.data.json';
	panel.innerHTML='<div class="form-grid" id="form-fields"><div class="empty"><p>Loading form...</p></div></div><div class="editor-toolbar"><span style="flex:1;font-size:.85rem;color:var(--muted)">Editing: '+source+'</span><button class="btn btn-primary btn-sm" id="form-save-btn" data-source="'+source+'" onclick="saveForm()">Save Changes</button></div>';

	fetch('/api/sites/'+slug+'/files/'+source).then(r=>r.json()).then(d=>{
		let vals={};
		try{vals=JSON.parse(d.content);}catch(e){}
		const ff=document.getElementById('form-fields');
		let html='';
		fields.forEach(f=>{
			const v=(vals[f.key]||'').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
			if(f.type==='textarea')html+='<div><label>'+f.label+'</label><textarea id="f_'+f.key+'">'+v+'</textarea></div>';
			else html+='<div><label>'+f.label+'</label><input type="'+(f.type==='number'?'number':'text')+'" id="f_'+f.key+'" value="'+v+'"></div>';
		});
		ff.innerHTML=html;
	}).catch(function(){document.getElementById('form-fields').innerHTML='<div class="empty"><p>Failed to load form data.</p></div>';});
}

function saveForm(){
	var btn=document.getElementById('form-save-btn');
	var source=btn.getAttribute('data-source');
	var fields=document.querySelectorAll('#form-fields input, #form-fields textarea');
	if(!fields.length){toast('Form not loaded yet, try again');return;}
	var obj={};
	fields.forEach(function(f){obj[f.id.replace('f_','')]=f.value;});
	btn.textContent='Saving...';
	fetch('/api/sites/'+slug+'/write',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({file:source,content:JSON.stringify(obj,null,'  ')})})
	.then(function(r){return r.json();}).then(function(d){
		if(d.success){toast('Saved!');btn.textContent='Save Changes';}
		else{toast('Error saving');btn.textContent='Save Changes';}
	}).catch(function(){toast('Network error');btn.textContent='Save Changes';});
}

function checkAuth(){var c=document.cookie.match('(^|; )sl_admin_session=([^;]*)');if(c){document.getElementById('auth-state').textContent='Logged in';}}
function logout(){location.href='/admin/logout?slug='+slug;}

buildUI();
checkAuth();
</script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// handleAdminAPIFiles lists files for the editor
func (s *Server) handleAdminAPIFiles(w http.ResponseWriter, r *http.Request, siteSlug string) {
	// Get site
	siteRepo := db.NewSiteRepository()
	site, err := siteRepo.GetBySlug(siteSlug)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Get path parameter
	queryPath := r.URL.Query().Get("path")

	// Get active version
	cfg, _ := config.Load()
	siteService := services.NewSiteService(cfg)
	version, err := siteService.GetVersionBySiteAndVersion(site.ID, "")
	if err != nil {
		// Try to get active version
		versionRepo := db.NewSiteVersionRepository()
		version, err = versionRepo.GetActiveVersion(site.ID)
		if err != nil {
			http.Error(w, "No active version", http.StatusNotFound)
			return
		}
	}

	// Determine directory to list
	dirPath := filepath.Join(cfg.SitesDir, site.Slug, version.Version, queryPath)

	// Read directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	var files []map[string]interface{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		isMarkdown := strings.HasSuffix(entry.Name(), ".md")
		files = append(files, map[string]interface{}{
			"name":       entry.Name(),
			"path":       filepath.Join(queryPath, entry.Name()),
			"is_markdown": isMarkdown,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"files": files})
}

// handleAdminAPIFileRead reads a file for the editor
func (s *Server) handleAdminAPIFileRead(w http.ResponseWriter, r *http.Request, siteSlug, filePath string) {
	// Get site
	siteRepo := db.NewSiteRepository()
	site, err := siteRepo.GetBySlug(siteSlug)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Get active version
	cfg, _ := config.Load()
	versionRepo := db.NewSiteVersionRepository()
	version, err := versionRepo.GetActiveVersion(site.ID)
	if err != nil {
		http.Error(w, "No active version", http.StatusNotFound)
		return
	}

	// Read file
	fullPath := filepath.Join(cfg.SitesDir, site.Slug, version.Version, filePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"content": string(content),
		"is_markdown": strings.HasSuffix(filePath, ".md"),
	})
}

// handleAdminAPIFileDelete deletes a file from the active version
func (s *Server) handleAdminAPIFileDelete(w http.ResponseWriter, r *http.Request, siteSlug, filePath string) {
	siteRepo := db.NewSiteRepository()
	site, err := siteRepo.GetBySlug(siteSlug)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	cfg, _ := config.Load()
	versionRepo := db.NewSiteVersionRepository()
	version, err := versionRepo.GetActiveVersion(site.ID)
	if err != nil {
		http.Error(w, "No active version", http.StatusNotFound)
		return
	}

	fullPath := filepath.Join(cfg.SitesDir, site.Slug, version.Version, filePath)
	if err := os.Remove(fullPath); err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleAPIUsers handles user API operations
func (s *Server) handleAPIUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// List users
		userRepo := db.NewUserRepository()
		users, err := userRepo.List()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"users": users})
		return
	}

	if r.Method == "POST" {
		// Create user
		var payload struct {
			Email    string `json:"email"`
			Password string `json:"password"`
			Role     string `json:"role"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		userRepo := db.NewUserRepository()
		user := &db.User{
			ID:   generateID(),
			Email: payload.Email,
			Role: payload.Role,
		}

		if err := userRepo.Create(user, payload.Password); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"user":    user,
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleAPIUserPassword handles password update
func (s *Server) handleAPIUserPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract email from path
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	email := strings.TrimSuffix(path, "/password")

	var payload struct {
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userRepo := db.NewUserRepository()
	if err := userRepo.UpdatePassword(email, payload.Password); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Password updated successfully",
	})
}

// handleAPIUserGrant handles granting site access
func (s *Server) handleAPIUserGrant(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload struct {
		SiteSlug string `json:"site_slug"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get site
	siteRepo := db.NewSiteRepository()
	site, err := siteRepo.GetBySlug(payload.SiteSlug)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Get user
	userRepo := db.NewUserRepository()
	user, err := userRepo.GetByEmail(payload.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Grant access
	if err := userRepo.GrantSiteAccess(site.ID, user.ID, payload.Role); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Granted %s access to %s", payload.Role, payload.SiteSlug),
	})
}

func generateRandomToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}

// handleAPISiteAdmin handles site admin token operations
func (s *Server) handleAPISiteAdmin(w http.ResponseWriter, r *http.Request) {
	// Extract site slug from path
	path := strings.TrimPrefix(r.URL.Path, "/sites/")
	parts := strings.Split(path, "/")
	siteSlug := parts[0]

	// Get site
	siteRepo := db.NewSiteRepository()
	site, err := siteRepo.GetBySlug(siteSlug)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	adminRepo := db.NewSiteAdminRepository()

	if r.Method == "POST" {
		// Create admin token
		token := generateRandomToken(32)
		expiresAt := time.Now().Add(30 * 24 * time.Hour)

		if err := adminRepo.CreateAdminToken(site.ID, token, &expiresAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		adminURL := fmt.Sprintf("/admin/%s/%s", siteSlug, token)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"admin_url": adminURL,
			"token":     token,
			"expires_at": expiresAt,
		})
		return
	}

	if r.Method == "GET" {
		// View admin token
		token, err := adminRepo.GetActiveTokenBySite(site.ID)
		if err != nil {
			http.Error(w, "No active admin token found", http.StatusNotFound)
			return
		}

		adminURL := fmt.Sprintf("/admin/%s/%s", siteSlug, token.Token)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"admin_url": adminURL,
			"token":     token.Token,
			"created_at": token.CreatedAt,
			"expires_at": token.ExpiresAt,
		})
		return
	}

	if r.Method == "PUT" {
		// Rotate admin token
		newToken := generateRandomToken(32)
		expiresAt := time.Now().Add(30 * 24 * time.Hour)

		if err := adminRepo.RotateToken(site.ID, newToken, &expiresAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		adminURL := fmt.Sprintf("/admin/%s/%s", siteSlug, newToken)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"admin_url": adminURL,
			"token":     newToken,
			"expires_at": expiresAt,
		})
		return
	}

	if r.Method == "DELETE" {
		// Revoke admin tokens
		if err := adminRepo.RevokeAllTokens(site.ID); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "All admin tokens revoked",
		})
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// JWT Claims
type Claims struct {
	UserID string `json:"user_id"`
	SiteID string `json:"site_id"`
	jwt.RegisteredClaims
}

// createJWT creates a JWT token
func createJWT(userID, siteID string, expiration time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		SiteID: siteID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// validateJWT validates a JWT token
func validateJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}