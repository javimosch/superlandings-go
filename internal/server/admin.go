package server

import (
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
	<style>
		* { box-sizing: border-box; }
		body { font-family: system-ui, sans-serif; margin: 0; padding: 0; background: #f5f5f5; }
		header { background: #007bff; color: white; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center; }
		h1 { margin: 0; font-size: 1.5rem; }
		.logout { background: rgba(255,255,255,0.2); color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; }
		.logout:hover { background: rgba(255,255,255,0.3); }
		main { padding: 2rem; max-width: 1400px; margin: 0 auto; }
		.card { background: white; border-radius: 8px; padding: 1.5rem; margin-bottom: 1rem; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		h2 { margin-top: 0; color: #333; }
		.file-list { list-style: none; padding: 0; }
		.file-list li { padding: 0.75rem; border-bottom: 1px solid #eee; cursor: pointer; display: flex; justify-content: space-between; }
		.file-list li:hover { background: #f9f9f9; }
		.file-list li:last-child { border-bottom: none; }
		.btn { background: #007bff; color: white; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; }
		.btn:hover { background: #0056b3; }
		.btn-secondary { background: #6c757d; }
		.btn-secondary:hover { background: #545b62; }
		.btn-success { background: #28a745; }
		.btn-success:hover { background: #218838; }
		.editor-container { display: none; margin-top: 1rem; }
		.editor-container.active { display: block; }
		textarea { width: 100%; min-height: 400px; font-family: monospace; padding: 1rem; border: 1px solid #ddd; border-radius: 4px; }
		.editor-actions { margin-top: 1rem; display: flex; gap: 0.5rem; }
		.modal { display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); }
		.modal.active { display: flex; justify-content: center; align-items: center; }
		.modal-content { background: white; padding: 2rem; border-radius: 8px; width: 100%; max-width: 500px; }
		.modal-content input { width: 100%; padding: 0.5rem; margin: 0.5rem 0; border: 1px solid #ddd; border-radius: 4px; }
		.modal-actions { margin-top: 1rem; display: flex; gap: 0.5rem; justify-content: flex-end; }
		.tabs { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
		.tab { background: #e9ecef; border: none; padding: 0.5rem 1rem; border-radius: 4px; cursor: pointer; }
		.tab.active { background: #007bff; color: white; }
		.tab-content { display: none; }
		.tab-content.active { display: block; }
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
			<textarea id="editor-content"></textarea>
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

		function showTab(tab) {
			document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
			document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
			event.target.classList.add('active');
			document.getElementById(tab + '-tab').classList.add('active');
			loadFiles(tab);
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
						li.innerHTML = '<span>' + f.name + '</span><button class="btn" onclick="editFile(\'' + f.path + '\', ' + (f.is_markdown || false) + ')">Edit</button>';
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
					document.getElementById('editor-content').value = data.content;
					document.getElementById('editor').classList.add('active');
				});
		}

		function saveFile() {
			const content = document.getElementById('editor-content').value;
			fetch('/api/sites/` + site.Slug + `/write', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ file: currentFile, content: content })
			}).then(r => r.json()).then(data => {
				if (data.success) {
					alert('Saved!');
					closeEditor();
					loadFiles('files');
					loadFiles('pages');
					loadFiles('blog');
				}
			});
		}

		function closeEditor() {
			document.getElementById('editor').classList.remove('active');
			currentFile = null;
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
func (s *Server) handleAdminAPIFiles(w http.ResponseWriter, r *http.Request) {
	// Extract site slug from path
	path := strings.TrimPrefix(r.URL.Path, "/api/sites/")
	parts := strings.Split(path, "/")
	siteSlug := parts[0]

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
func (s *Server) handleAdminAPIFileRead(w http.ResponseWriter, r *http.Request) {
	// Extract site slug and file path from path
	path := strings.TrimPrefix(r.URL.Path, "/api/sites/")
	parts := strings.SplitN(path, "/", 3)
	if len(parts) < 3 {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	siteSlug := parts[0]
	filePath := parts[2]

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