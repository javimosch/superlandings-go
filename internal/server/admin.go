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

	if len(parts) < 2 {
		http.Error(w, "Invalid admin URL", http.StatusBadRequest)
		return
	}

	siteSlug := parts[0]
	token := parts[1]

	// Verify token
	adminRepo := db.NewSiteAdminRepository()
	adminToken, err := adminRepo.GetTokenByValue(token)
	if err != nil {
		http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
		return
	}

	if !adminToken.IsActive {
		http.Error(w, "Token has been revoked", http.StatusUnauthorized)
		return
	}

	// Check expiration
	if adminToken.ExpiresAt != nil && time.Now().After(*adminToken.ExpiresAt) {
		http.Error(w, "Token has expired", http.StatusUnauthorized)
		return
	}

	// Get site
	siteRepo := db.NewSiteRepository()
	site, err := siteRepo.GetBySlug(siteSlug)
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	// Token is valid — go directly to editor (no login required)
	s.handleAdminEditor(w, r, site)
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

		// Redirect to editor
		http.Redirect(w, r, "/admin/"+site.Slug+"/"+r.URL.Path[len("/admin/"):], http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// handleAdminEditor serves the editor UI
func (s *Server) handleAdminEditor(w http.ResponseWriter, r *http.Request, site *db.Site) {

	html := `<!DOCTYPE html>
<html>
<head>
	<title>` + site.Name + ` &mdash; Editor</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/easymde@2.18.0/dist/easymde.min.css">
	<script src="https://cdn.jsdelivr.net/npm/easymde@2.18.0/dist/easymde.min.js"></script>
	<style>
		:root{--primary:#2563eb;--accent:#7c3aed;--bg:#f8fafc;--card:#fff;--text:#1e293b;--muted:#94a3b8;--border:#e2e8f0}
		*{margin:0;padding:0;box-sizing:border-box}
		body{font-family:system-ui,-apple-system,sans-serif;background:var(--bg);color:var(--text);line-height:1.6}
		.hdr{background:var(--card);border-bottom:1px solid var(--border);padding:.75rem 1.5rem;display:flex;align-items:center;justify-content:space-between;position:sticky;top:0;z-index:10}
		.hdr h1{font-size:1.1rem;font-weight:600}
		.hdr .site{color:var(--muted);font-weight:400}
		.hdr a{color:var(--primary);text-decoration:none;font-size:.9rem}
		.wrap{display:flex;height:calc(100vh - 56px)}
		.sidebar{width:260px;background:var(--card);border-right:1px solid var(--border);padding:1rem;overflow-y:auto;flex-shrink:0}
		.sidebar h2{font-size:.75rem;text-transform:uppercase;letter-spacing:.05em;color:var(--muted);margin:1rem 0 .5rem}
		.post-list{list-style:none}
		.post-list li{padding:.5rem .75rem;border-radius:6px;cursor:pointer;font-size:.875rem;display:flex;align-items:center;justify-content:space-between;transition:background .15s}
		.post-list li:hover{background:#f1f5f9}
		.post-list li .tt{font-weight:500}
		.post-list li .tag{font-size:.65rem;background:#e0f2fe;color:#0284c7;padding:.15rem .35rem;border-radius:3px}
		.main{flex:1;display:flex;flex-direction:column;overflow:hidden}
		.empty{flex:1;display:flex;flex-direction:column;align-items:center;justify-content:center;color:var(--muted);padding:2rem;text-align:center}
		.empty h2{font-size:1.25rem;color:var(--text);margin-bottom:.5rem}
		.empty p{font-size:.9rem;margin-bottom:1.5rem;max-width:400px}
		.editor-area{flex:1;display:none;flex-direction:column;overflow:hidden}
		.editor-area.active{display:flex}
		.editor-toolbar{display:flex;align-items:center;gap:.75rem;padding:.75rem 1.5rem;border-bottom:1px solid var(--border);background:var(--card)}
		.editor-toolbar input{flex:1;border:none;font-size:1.1rem;font-weight:600;outline:none;background:transparent;color:var(--text)}
		.EasyMDEContainer{border:none!important;border-radius:0!important;flex:1;display:flex;flex-direction:column}
		.EasyMDEContainer .editor-toolbar{border:none!important;border-bottom:1px solid var(--border)!important}
		.EasyMDEContainer .CodeMirror{flex:1!important;border:none!important;border-radius:0!important;font-size:.95rem!important}
		.btn{display:inline-flex;align-items:center;gap:.4rem;padding:.5rem 1rem;border-radius:6px;font-size:.875rem;font-weight:500;border:none;cursor:pointer;transition:all .15s}
		.btn-primary{background:var(--primary);color:#fff}.btn-primary:hover{background:#1d4ed8}
		.btn-success{background:#059669;color:#fff}.btn-success:hover{background:#047857}
		.btn-outline{background:transparent;color:var(--text);border:1px solid var(--border)}.btn-outline:hover{background:#f1f5f9}
		.btn-sm{padding:.35rem .75rem;font-size:.8rem}
		.modal{display:none;position:fixed;inset:0;background:rgba(0,0,0,.5);z-index:100;align-items:center;justify-content:center}
		.modal.active{display:flex}
		.modal-inner{background:var(--card);padding:1.5rem;border-radius:12px;width:100%;max-width:400px}
		.modal-inner h2{font-size:1.1rem;margin-bottom:.75rem}
		.modal-inner input{width:100%;padding:.6rem .75rem;border:1px solid var(--border);border-radius:6px;font-size:.9rem;outline:none}
		.modal-actions{display:flex;gap:.5rem;justify-content:flex-end;margin-top:.75rem}
		.toast{position:fixed;bottom:1.5rem;right:1.5rem;background:#065f46;color:#fff;padding:.75rem 1.25rem;border-radius:8px;font-size:.875rem;box-shadow:0 4px 12px rgba(0,0,0,.15);opacity:0;transform:translateY(10px);transition:all .3s;z-index:200}
		.toast.show{opacity:1;transform:translateY(0)}
	</style>
</head>
<body>
<div class="hdr">
	<h1><span class="site">` + site.Name + `</span> Editor</h1>
	<a href="/` + site.Slug + `" target="_blank">View site &rarr;</a>
</div>
<div class="wrap">
	<div class="sidebar">
		<h2>Blog Posts</h2>
		<ul class="post-list" id="post-list"></ul>
		<button class="btn btn-success btn-sm" style="width:100%;margin-top:.5rem" onclick="newPost()">+ New Post</button>
	</div>
	<div class="main">
		<div class="empty" id="empty-state">
			<h2>Welcome</h2>
			<p>Write a new blog post or select one from the sidebar.</p>
			<button class="btn btn-success" onclick="newPost()">+ Write Your First Post</button>
		</div>
		<div class="editor-area" id="editor-area">
			<div class="editor-toolbar">
				<input type="text" id="post-title" placeholder="Post title...">
				<button class="btn btn-primary btn-sm" onclick="savePost()">Publish</button>
			</div>
			<textarea id="markdown-editor"></textarea>
		</div>
	</div>
</div>
<div class="modal" id="modal"><div class="modal-inner">
	<h2>New Post</h2>
	<p style="font-size:.875rem;color:var(--muted);margin-bottom:.75rem">Enter a URL slug:</p>
	<input type="text" id="modal-input" placeholder="my-new-article">
	<div class="modal-actions">
		<button class="btn btn-primary" onclick="confirmNewPost()">Create</button>
		<button class="btn btn-outline" onclick="closeModal()">Cancel</button>
	</div>
</div></div>
<div class="toast" id="toast">Saved!</div>
<script>
let currentPost=null,easyMDE=null,modalCallback=null;
const slug="` + site.Slug + `";
function toast(m){const t=document.getElementById('toast');t.textContent=m;t.classList.add('show');setTimeout(()=>t.classList.remove('show'),2000)}
function loadPosts(){
	fetch('/api/sites/'+slug+'/files?path=blog').then(r=>r.json()).then(d=>{
		const list=document.getElementById('post-list');list.innerHTML='';
		(d.files||[]).forEach(f=>{
			const n=f.name.replace(/\.md$/,'').replace(/-/g,' ');const lbl=n.charAt(0).toUpperCase()+n.slice(1);
			const li=document.createElement('li');
			li.innerHTML='<span class="tt">'+lbl+'</span><span class="tag">md</span>';
			li.onclick=()=>editPost(f.path);list.appendChild(li);
		});
	});
}
function editPost(path){
	currentPost=path;
	fetch('/api/sites/'+slug+'/files/'+path).then(r=>r.json()).then(d=>{
		document.getElementById('empty-state').style.display='none';document.getElementById('editor-area').classList.add('active');
		const lines=d.content.split('\n');let title='',body=d.content;
		for(const l of lines){if(l.startsWith('# ')){title=l.replace(/^# /,'').trim();body=lines.slice(lines.indexOf(l)+1).join('\n').trim();break;}}
		document.getElementById('post-title').value=title||'Untitled';
		if(easyMDE){easyMDE.value(body);}else{
			easyMDE=new EasyMDE({element:document.getElementById('markdown-editor'),spellChecker:false,autofocus:true,placeholder:'Write your post...',status:false,toolbar:['bold','italic','heading','|','quote','unordered-list','ordered-list','|','link','image','|','preview','side-by-side','fullscreen']});
			easyMDE.value(body);
		}
	});
}
function savePost(){
	const title=document.getElementById('post-title').value.trim();const body=easyMDE?easyMDE.value().trim():'';
	if(!title&&!body){toast('Nothing to save');return;}
	const content=(title?'# '+title+'\n\n':'')+body;
	if(currentPost){
		fetch('/api/sites/'+slug+'/write',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({file:currentPost,content:content})})
		.then(r=>r.json()).then(d=>{if(d.success){toast('Published!');loadPosts();}});
	}else{
		const slugName=title.toLowerCase().replace(/[^a-z0-9]+/g,'-').replace(/^-|-$/g,'')||'untitled';
		const fp='blog/'+slugName+'.md';
		fetch('/api/sites/'+slug+'/write',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({file:fp,content:content})})
		.then(r=>r.json()).then(d=>{if(d.success){currentPost=fp;toast('Published! /'+slug+'/'+slugName);loadPosts();}});
	}
}
function newPost(){if(easyMDE)easyMDE.value('');document.getElementById('post-title').value='';document.getElementById('empty-state').style.display='none';document.getElementById('editor-area').classList.add('active');currentPost=null;setTimeout(()=>document.getElementById('post-title').focus(),100);}
loadPosts();
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