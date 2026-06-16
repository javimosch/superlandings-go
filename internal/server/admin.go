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

	// Check if user is already logged in
	sessionCookie, err := r.Cookie("sl_admin_session")
	if err == nil {
		// Validate session
		claims, err := validateJWT(sessionCookie.Value)
		if err == nil && claims.SiteID == site.ID {
			// Valid session, show editor
			s.handleAdminEditor(w, r, site)
			return
		}
	}

	// Not logged in, show login form
	s.handleAdminLogin(w, r, site)
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
	<title>Editor - ` + site.Name + `</title>
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/easymde@2.18.0/dist/easymde.min.css">
	<script src="https://cdn.jsdelivr.net/npm/easymde@2.18.0/dist/easymde.min.js"></script>
	<style>
		* { box-sizing: border-box; }
		body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; margin: 0; padding: 0; background: #f8fafc; }
		header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center; box-shadow: 0 4px 6px rgba(0,0,0,0.1); }
		h1 { margin: 0; font-size: 1.5rem; font-weight: 700; }
		.logout { background: rgba(255,255,255,0.2); color: white; border: none; padding: 0.5rem 1rem; border-radius: 6px; cursor: pointer; font-weight: 500; transition: background 0.2s; }
		.logout:hover { background: rgba(255,255,255,0.3); }
		main { padding: 2rem; max-width: 1400px; margin: 0 auto; }
		.card { background: white; border-radius: 12px; padding: 1.5rem; margin-bottom: 1.5rem; box-shadow: 0 4px 6px rgba(0,0,0,0.05), 0 1px 3px rgba(0,0,0,0.1); border: 1px solid #e5e7eb; }
		h2 { margin-top: 0; color: #1f2937; font-weight: 600; font-size: 1.25rem; }
		.file-list { list-style: none; padding: 0; margin: 0; }
		.file-list li { padding: 1rem; border-bottom: 1px solid #f3f4f6; cursor: pointer; display: flex; align-items: center; justify-content: space-between; border-radius: 8px; transition: all 0.2s; margin-bottom: 0.5rem; }
		.file-list li:hover { background: #f9fafb; transform: translateX(4px); }
		.file-list li:last-child { border-bottom: none; }
		.file-item { display: flex; align-items: center; gap: 0.75rem; flex: 1; }
		.file-icon { width: 32px; height: 32px; border-radius: 6px; display: flex; align-items: center; justify-content: center; font-size: 14px; }
		.file-icon.html { background: #e0f2fe; color: #0284c7; }
		.file-icon.md { background: #fef3c7; color: #d97706; }
		.file-icon.json { background: #dcfce7; color: #16a34a; }
		.file-name { font-weight: 500; color: #374151; }
		.file-path { font-size: 0.875rem; color: #6b7280; }
		.btn { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; border: none; padding: 0.5rem 1rem; border-radius: 6px; cursor: pointer; font-weight: 500; transition: all 0.2s; }
		.btn:hover { opacity: 0.9; transform: translateY(-1px); }
		.btn-secondary { background: #6b7280; }
		.btn-secondary:hover { background: #4b5563; }
		.btn-success { background: #10b981; }
		.btn-success:hover { background: #059669; }
		.editor-container { display: none; margin-top: 1.5rem; }
		.editor-container.active { display: block; }
		textarea { width: 100%; min-height: 400px; font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace; padding: 1rem; border: 1px solid #d1d5db; border-radius: 8px; font-size: 14px; line-height: 1.6; }
		.editor-actions { margin-top: 1rem; display: flex; gap: 0.75rem; }
		.modal { display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); backdrop-filter: blur(4px); }
		.modal.active { display: flex; justify-content: center; align-items: center; }
		.modal-content { background: white; padding: 2rem; border-radius: 12px; width: 100%; max-width: 500px; box-shadow: 0 20px 25px rgba(0,0,0,0.1); }
		.modal-content input { width: 100%; padding: 0.75rem; margin: 0.75rem 0; border: 1px solid #d1d5db; border-radius: 6px; font-size: 1rem; }
		.modal-content input:focus { outline: none; border-color: #667eea; box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1); }
		.modal-actions { margin-top: 1rem; display: flex; gap: 0.75rem; justify-content: flex-end; }
		.tabs { display: flex; gap: 0.5rem; margin-bottom: 1.5rem; background: white; padding: 0.5rem; border-radius: 8px; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
		.tab { background: transparent; border: none; padding: 0.5rem 1rem; border-radius: 6px; cursor: pointer; font-weight: 500; color: #6b7280; transition: all 0.2s; }
		.tab:hover { background: #f3f4f6; }
		.tab.active { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; }
		.tab-content { display: none; }
		.tab-content.active { display: block; }
		.EasyMDEContainer { border: 1px solid #d1d5db; border-radius: 8px; }
		.editor-wrapper { display: none; }
		.editor-wrapper.active { display: block; }
		.markdown-editor { display: none; }
		.markdown-editor.active { display: block; }
	</style>
</head>
<body>
	<header>
		<h1>Editor: ` + site.Name + `</h1>
		<button class="logout" onclick="logout()">Logout</button>
	</header>
	<main>
		<div class="tabs">
			<button class="tab active" onclick="showTab('files')">Files</button>
			<button class="tab" onclick="showTab('pages')">Pages</button>
			<button class="tab" onclick="showTab('blog')">Blog</button>
		</div>

		<div id="files-tab" class="tab-content active">
			<div class="card">
				<h2>Files</h2>
				<button class="btn btn-success" onclick="showCreateFileModal('')">+ New File</button>
				<ul class="file-list" id="file-list"></ul>
			</div>
		</div>

		<div id="pages-tab" class="tab-content">
			<div class="card">
				<h2>Pages</h2>
				<button class="btn btn-success" onclick="showCreateFileModal('pages/')">+ New Page</button>
				<ul class="file-list" id="page-list"></ul>
			</div>
		</div>

		<div id="blog-tab" class="tab-content">
			<div class="card">
				<h2>Blog Posts</h2>
				<button class="btn btn-success" onclick="showCreateFileModal('blog/', true)">+ New Post</button>
				<ul class="file-list" id="blog-list"></ul>
			</div>
		</div>

		<div class="card editor-container" id="editor">
			<h2 id="editor-title">Edit File</h2>
			<div id="plain-editor" class="editor-wrapper active">
				<textarea id="editor-content"></textarea>
			</div>
			<div id="markdown-editor" class="markdown-editor">
				<textarea id="markdown-content"></textarea>
			</div>
			<div class="editor-actions">
				<button class="btn" onclick="saveFile()">Save</button>
				<button class="btn btn-secondary" onclick="closeEditor()">Cancel</button>
			</div>
		</div>
	</main>

	<div class="modal" id="create-modal">
		<div class="modal-content">
			<h2 id="modal-title">Create File</h2>
			<input type="text" id="new-file-name" placeholder="File name (e.g., about-me.html or my-post.md)">
			<div class="modal-actions">
				<button class="btn" onclick="createFile()">Create</button>
				<button class="btn btn-secondary" onclick="closeModal()">Cancel</button>
			</div>
		</div>
	</div>

	<script>
		let currentFile = null;
		let currentPath = '';
		let isMarkdown = false;
		let easyMDE = null;

		function showTab(tab) {
			document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
			document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
			event.target.classList.add('active');
			document.getElementById(tab + '-tab').classList.add('active');
			loadFiles(tab);
		}

		function getFileIcon(name, isMarkdown) {
			if (name.endsWith('.md')) return '<div class="file-icon md">📝</div>';
			if (name.endsWith('.json')) return '<div class="file-icon json">{ }</div>';
			return '<div class="file-icon html">🌐</div>';
		}

		function loadFiles(type) {
			const path = type === 'files' ? '' : type;
			fetch('/api/sites/` + site.Slug + `/files?path=' + path)
				.then(r => r.json())
				.then(data => {
					const listId = type === 'files' ? 'file-list' : (type === 'pages' ? 'page-list' : 'blog-list');
					const list = document.getElementById(listId);
					list.innerHTML = '';
					data.files.forEach(f => {
						const li = document.createElement('li');
						li.onclick = () => editFile(f.path, f.is_markdown || false);
						li.innerHTML = '<div class="file-item">' + getFileIcon(f.name, f.is_markdown) + '<div><div class="file-name">' + f.name + '</div><div class="file-path">' + f.path + '</div></div></div>';
						list.appendChild(li);
					});
				});
		}

		function editFile(path, isMd) {
			currentFile = path;
			isMarkdown = isMd;
			fetch('/api/sites/` + site.Slug + `/files/' + path)
				.then(r => r.json())
				.then(data => {
					document.getElementById('editor-title').textContent = 'Edit: ' + path;
					document.getElementById('editor').classList.add('active');
					
					if (isMd) {
						document.getElementById('plain-editor').classList.remove('active');
						document.getElementById('markdown-editor').classList.add('active');
						document.getElementById('markdown-content').value = data.content;
						
						if (easyMDE) {
							easyMDE.value(data.content);
						} else {
							easyMDE = new EasyMDE({
								element: document.getElementById('markdown-content'),
								spellChecker: false,
								autofocus: true,
								placeholder: 'Write your markdown here...',
								status: false,
								toolbar: ['bold', 'italic', 'heading', '|', 'quote', 'unordered-list', 'ordered-list', '|', 'link', 'image', '|', 'preview', 'side-by-side', 'fullscreen']
							});
						}
					} else {
						document.getElementById('markdown-editor').classList.remove('active');
						document.getElementById('plain-editor').classList.add('active');
						document.getElementById('editor-content').value = data.content;
					}
				});
		}

		function saveFile() {
			let content;
			if (isMarkdown && easyMDE) {
				content = easyMDE.value();
			} else if (isMarkdown) {
				content = document.getElementById('markdown-content').value;
			} else {
				content = document.getElementById('editor-content').value;
			}
			
			fetch('/api/sites/` + site.Slug + `/write', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ file: currentFile, content: content })
			}).then(r => r.json()).then(data => {
				if (data.success) {
					alert('Saved successfully!');
					closeEditor();
					loadFiles('files');
					loadFiles('pages');
					loadFiles('blog');
				}
			});
		}

		function closeEditor() {
			document.getElementById('editor').classList.remove('active');
			document.getElementById('plain-editor').classList.remove('active');
			document.getElementById('markdown-editor').classList.remove('active');
			currentFile = null;
			isMarkdown = false;
		}

		function showCreateFileModal(path, isMd = false) {
			currentPath = path;
			isMarkdown = isMd;
			document.getElementById('modal-title').textContent = isMd ? 'Create Blog Post' : 'Create File';
			document.getElementById('new-file-name').placeholder = isMd ? 'my-post.md' : 'file.html';
			document.getElementById('create-modal').classList.add('active');
		}

		function closeModal() {
			document.getElementById('create-modal').classList.remove('active');
			document.getElementById('new-file-name').value = '';
		}

		function createFile() {
			const name = document.getElementById('new-file-name').value;
			if (!name) return;
			const path = currentPath + name;
			fetch('/api/sites/` + site.Slug + `/write', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ file: path, content: '' })
			}).then(r => r.json()).then(data => {
				if (data.success) {
					closeModal();
					loadFiles(currentPath ? currentPath.replace(/\/$/, '') : 'files');
					editFile(path, isMarkdown);
				}
			});
		}

		function logout() {
			document.cookie = 'sl_admin_session=; path=/admin/` + site.Slug + `; expires=Thu, 01 Jan 1970 00:00:00 GMT';
			window.location.reload();
		}

		// Load files on init
		loadFiles('files');
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