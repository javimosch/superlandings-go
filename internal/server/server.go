package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/javimosch/superlandings-go/internal/services"
)

type Server struct {
	cfg         *config.Config
	siteService *services.SiteService
	dnsService  *services.DNSService
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg:         cfg,
		siteService: services.NewSiteService(cfg),
		dnsService:  services.NewDNSService(cfg),
	}
}

// Start starts the HTTP server
func (s *Server) Start(port int) error {
	// Initialize database once and keep it open
	if err := db.Initialize(s.cfg.DatabasePath); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	// Note: Database stays open for server lifetime

	// Setup routes
	mux := http.NewServeMux()

	// Admin routes must be registered before the catch-all / handler
	mux.HandleFunc("/admin/logout", s.handleAdminLogout)
	mux.HandleFunc("/admin/", s.handleAdmin)

	// API routes must be registered before the catch-all / handler
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/status", s.authMiddleware(s.handleAPIStatus))
	apiMux.HandleFunc("/users", s.authMiddleware(s.handleAPIUsers))
	apiMux.HandleFunc("/users/", s.authMiddleware(s.handleAPIUserPassword))
	apiMux.HandleFunc("/users/grant", s.authMiddleware(s.handleAPIUserGrant))
	apiMux.HandleFunc("/sites", s.authMiddleware(s.handleAPISites))
	apiMux.HandleFunc("/sites/", s.authMiddleware(s.handleAPISite))

	mux.Handle("/api/", http.StripPrefix("/api", apiMux))
	mux.HandleFunc("/", s.handleLanding)
	mux.HandleFunc("/health", s.handleHealth)

	// Start server
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Server starting on http://localhost%s", addr)
	log.Printf("Landings will be served at http://localhost%s/:slug", addr)

	return http.ListenAndServe(addr, mux)
}

// handleLanding serves landing pages and sites
func (s *Server) handleLanding(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	// Query param fallback for local testing: ?site=vdb-landing
	// ?site=clear removes the cookie and shows the landing page list
	if qs := r.URL.Query().Get("site"); qs != "" {
		if qs == "clear" {
			http.SetCookie(w, &http.Cookie{Name: "sl_site", Value: "", Path: "/", MaxAge: -1})
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:  "sl_site",
			Value: qs,
			Path:  "/",
		})
		http.Redirect(w, r, "/"+path, http.StatusFound)
		return
	}

	if path == "" {
		// Try host-based resolution for root path
		siteSlug, _, fromDomain := s.resolveSite("/", r.Host)
		if fromDomain && siteSlug != "" {
			if content, err := s.siteService.GetActiveVersionContent(siteSlug, ""); err == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(content))
				return
			}
		}
		// Cookie-based fallback for local testing
		if c, err := r.Cookie("sl_site"); err == nil && c.Value != "" {
			if content, err := s.siteService.GetActiveVersionContent(c.Value, ""); err == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(content))
				return
			}
		}
		s.handleRoot(w, r)
		return
	}

	// Resolve site slug: try path prefix first, then Host header (domain-based)
	siteSlug, filePath, fromDomain := s.resolveSite(path, r.Host)

	// Cookie-based fallback for local testing (set by ?site= param)
	if _, err := s.siteService.GetBySlug(siteSlug); err != nil {
		if c, cErr := r.Cookie("sl_site"); cErr == nil && c.Value != "" {
			siteSlug = c.Value
			filePath = path
			fromDomain = true
		}
	}

	// Try to serve as a static asset first (shared across versions)
	if filePath != "" && isAssetExt(filePath) {
		if served := s.tryServeAsset(w, r, siteSlug, filePath); served {
			return
		}
	}

	// Try to serve as a site (with dynamic blocks and sub-paths)
	if content, err := s.siteService.GetActiveVersionContent(siteSlug, filePath); err == nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(content))
		return
	}

	// If resolved from domain, the landing fallback doesn't apply
	if fromDomain {
		http.NotFound(w, r)
		return
	}

	// Not a known site or domain
	http.NotFound(w, r)
}

// handleRoot serves the root page
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	sites, err := s.siteService.List()
	if err != nil {
		http.Error(w, "Failed to list sites", http.StatusInternalServerError)
		return
	}

	html := `<!DOCTYPE html><html><head><title>SuperLandings</title><meta name="viewport" content="width=device-width,initial-scale=1"><style>
		body{font-family:system-ui,sans-serif;background:#0a0a0f;color:#e4e4ec;max-width:720px;margin:3rem auto;padding:0 1.5rem}
		h1{font-size:1.5rem;font-weight:700;margin-bottom:.25rem}
		h1 span{color:#6c5ce7}
		.sub{color:#7c7c94;font-size:.9rem;margin-bottom:2rem}
		.section{margin-bottom:2rem}
		h2{font-size:.85rem;text-transform:uppercase;letter-spacing:.05em;color:#7c7c94;margin-bottom:.75rem}
		a{color:#a29bfe;text-decoration:none;display:block;padding:.6rem .75rem;border-radius:6px;transition:background .15s}
		a:hover{background:#1a1a2e}
		.row{display:flex;align-items:center;justify-content:space-between}
		.root-btn{font-size:.7rem;background:#6c5ce7;color:#fff;border:none;border-radius:4px;padding:.2rem .5rem;cursor:pointer}
		.root-btn:hover{background:#a29bfe}
	</style></head><body>
	<h1>Super<span>Landings</span></h1>
	<p class="sub">` + fmt.Sprintf("%d sites", len(sites)) + ` &middot; <code style="font-size:.8rem">sl-cli site create</code> to add</p>
	<div class="section"><h2>Sites</h2>`

	for _, site := range sites {
		html += fmt.Sprintf(`<div class="row"><a href="/%s">%s</a><button class="root-btn" onclick="navigator.clipboard.writeText('%s/?site=%s').then(()=>this.textContent='Copied!')">Open at root</button></div>`,
			site.Slug, site.Name, r.Host, site.Slug)
	}
	if len(sites) == 0 {
		html += `<p style="color:#7c7c94">No sites yet. Create one: <code>sl-cli site create --name "My Site" --slug "my-site"</code></p>`
	}

	html += `</div></body></html>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// handleHealth serves health check
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"healthy"}`))
}

// API handlers for remote execution
func (s *Server) handleAPIStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"running","service":"sl-cli-daemon"}`))
}

func (s *Server) handleAPISites(w http.ResponseWriter, r *http.Request) {
	sites, err := s.siteService.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	
	dnsService := services.NewDNSService(s.cfg)

	// Convert to JSON manually to avoid extra dependencies
	json := "{\"sites\":["
	for i, site := range sites {
		if i > 0 {
			json += ","
		}

		// Get domains for this site
		domains, _ := dnsService.GetDomains(site.ID)
		domainList := "[]"
		if len(domains) > 0 {
			domainList = "["
			for j, d := range domains {
				if j > 0 {
					domainList += ","
				}
				domainList += fmt.Sprintf(`"%s"`, d.Domain)
			}
			domainList += "]"
		}

		json += fmt.Sprintf(`{"id":"%s","slug":"%s","name":"%s","domains":%s}`,
			site.ID, site.Slug, site.Name, domainList)
	}
	json += "]}"
	w.Write([]byte(json))
}

func (s *Server) handleAPISite(w http.ResponseWriter, r *http.Request) {
	// Extract site slug from path
	// Path format: /sites/{slug} or /sites/{slug}/{action} (after /api/ is stripped)
	path := strings.TrimPrefix(r.URL.Path, "/sites/")
	parts := strings.Split(path, "/")
	slug := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}
	
	// Check for nested actions (e.g., versions/switch)
	nestedAction := ""
	if len(parts) > 2 {
		nestedAction = parts[2]
	}
	
	if action == "versions" && nestedAction == "switch" {
		s.handleAPISiteVersionSwitch(w, r, slug)
		return
	}
	
	switch action {
	case "versions":
		s.handleAPISiteVersions(w, r, slug)
	case "sync":
		s.handleAPISiteSync(w, r, slug)
	case "dns":
		s.handleAPISiteDNS(w, r, slug)
	case "write":
		s.handleAPISiteWrite(w, r, slug)
	case "write-batch":
		s.handleAPISiteWriteBatch(w, r, slug)
	case "upload":
		s.handleAPISiteUpload(w, r, slug)
	case "assets":
		s.handleAPISiteAssets(w, r, slug)
	case "files":
		if len(parts) > 2 {
			filePath := strings.Join(parts[2:], "/")
			if r.Method == "DELETE" {
				s.handleAdminAPIFileDelete(w, r, slug, filePath)
				return
			}
			s.handleAdminAPIFileRead(w, r, slug, filePath)
			return
		}
		s.handleAdminAPIFiles(w, r, slug)
	case "admin":
		s.handleAPISiteAdmin(w, r)
	case "forms":
		s.handleAPISiteForms(w, r, slug, parts[2:])
	default:
		s.handleAPISiteDetails(w, r, slug)
	}
}

func (s *Server) handleAPISiteDetails(w http.ResponseWriter, r *http.Request, slug string) {
	sites, err := s.siteService.List()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Find site by slug
	var site *db.Site
	for _, s := range sites {
		if s.Slug == slug {
			site = &s
			break
		}
	}
	
	if site == nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json := fmt.Sprintf(`{"slug":"%s","name":"%s"}`,
		site.Slug, site.Name)
	w.Write([]byte(json))
}

func (s *Server) handleAPISiteVersions(w http.ResponseWriter, r *http.Request, slug string) {
	if r.Method == "GET" {
		versions, err := s.siteService.ListVersions(slug)
		if err != nil {
			http.Error(w, "Site not found", http.StatusNotFound)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json := "{\"versions\":["
		for i, v := range versions {
			if i > 0 {
				json += ","
			}
			json += fmt.Sprintf(`{"version":"%s","comment":"%s","is_active":%t,"path":"%s"}`,
				v.Version, v.Comment, v.IsActive, v.Path)
		}
		json += "]}"
		w.Write([]byte(json))
		return
	}
	
	if r.Method == "POST" {
		var payload struct {
			Version string `json:"version"`
			Comment string `json:"comment"`
			Author  string `json:"author"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		if payload.Version == "" {
			http.Error(w, "version is required", http.StatusBadRequest)
			return
		}
		
		req := services.CreateVersionRequest{
			Version: payload.Version,
			Comment: payload.Comment,
			Author:  payload.Author,
		}
		
		createdVersion, err := s.siteService.CreateVersion(slug, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json := fmt.Sprintf(`{"version":"%s","comment":"%s","is_active":%t,"path":"%s"}`,
			createdVersion.Version, createdVersion.Comment, createdVersion.IsActive, createdVersion.Path)
		w.Write([]byte(json))
		return
	}
	
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleAPISiteSync(w http.ResponseWriter, r *http.Request, slug string) {
	// Only handle POST requests
	if r.Method != "POST" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"success":false,"error":"method not allowed"}`))
		return
	}
	
	// Check if sync target is configured
	if s.cfg.SyncTargetHost == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"success":false,"error":"sync target not configured on daemon"}`))
		return
	}
	
	// Check if site exists
	sites, err := s.siteService.List()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success":false,"error":"failed to list sites"}`))
		return
	}
	
	siteExists := false
	for _, site := range sites {
		if site.Slug == slug {
			siteExists = true
			break
		}
	}
	
	if !siteExists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success":false,"error":"site not found"}`))
		return
	}
	
	// Trigger sync service
	syncService := services.NewSyncService(s.cfg)
	syncTarget := services.SyncTarget{
		Host: s.cfg.SyncTargetHost,
		User: s.cfg.SyncTargetUser,
		Port: s.cfg.SyncTargetPort,
		Key:  s.cfg.SyncTargetKey,
	}
	
	if err := syncService.Sync(slug, syncTarget); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"sync failed: %s"}`, err.Error())))
		return
	}
	
	// Return success
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true,"message":"site synced successfully"}`))
}

func (s *Server) handleAPISiteDNS(w http.ResponseWriter, r *http.Request, slug string) {
	// Get site by slug
	sites, err := s.siteService.List()
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}
	
	var site *db.Site
	for _, s := range sites {
		if s.Slug == slug {
			site = &s
			break
		}
	}
	
	if site == nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}
	
	dnsService := services.NewDNSService(s.cfg)
	
	if r.Method == "GET" {
		// List DNS entries
		domains, err := dnsService.GetDomains(site.ID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json := "{\"domains\":["
		for i, d := range domains {
			if i > 0 {
				json += ","
			}
			json += fmt.Sprintf(`{"domain":"%s","ip":"%s","traefik":%t}`,
				d.Domain, d.IP, d.Traefik)
		}
		json += "]}"
		w.Write([]byte(json))
		return
	}
	
	if r.Method == "POST" {
		// Parse request body
		var payload struct {
			Domain  string `json:"domain"`
			IP      string `json:"ip"`
			Traefik bool   `json:"traefik"`
			Action  string `json:"action"` // "setup" or "remove"
		}
		
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		// Determine action from URL path
		path := strings.TrimPrefix(r.URL.Path, "/sites/")
		parts := strings.Split(path, "/")
		action := ""
		if len(parts) > 2 {
			action = parts[2]
		}
		
		if action == "setup" {
			if payload.Domain == "" || payload.IP == "" {
				http.Error(w, "domain and ip are required", http.StatusBadRequest)
				return
			}
			
			if err := dnsService.SetupDNS(site.ID, slug, payload.Domain, payload.IP, payload.Traefik); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"success":true}`))
			return
		}
		
		if action == "remove" {
			// RemoveDNS removes all DNS for a site via hotify-cli prune
			if err := dnsService.RemoveDNS(slug); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"success":true}`))
			return
		}
		
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}
	
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (s *Server) handleAPISiteVersionSwitch(w http.ResponseWriter, r *http.Request, slug string) {
	var payload struct {
		Version string `json:"version"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if payload.Version == "" {
		http.Error(w, "version is required", http.StatusBadRequest)
		return
	}
	
	if err := s.siteService.SwitchVersion(slug, payload.Version); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

func (s *Server) handleAPISiteWrite(w http.ResponseWriter, r *http.Request, slug string) {
	var payload struct {
		Version string `json:"version"`
		File    string `json:"file"`
		Content string `json:"content"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if payload.File == "" {
		http.Error(w, "file is required", http.StatusBadRequest)
		return
	}
	if payload.Content == "" {
		http.Error(w, "content is required", http.StatusBadRequest)
		return
	}
	
	// Default to active version if not specified
	if payload.Version == "" {
		site, err := s.siteService.GetBySlug(slug)
		if err != nil {
			http.Error(w, "Site not found", http.StatusNotFound)
			return
		}
		
		versionRepo := db.NewSiteVersionRepository()
		version, err := versionRepo.GetActiveVersion(site.ID)
		if err != nil {
			http.Error(w, "No active version", http.StatusNotFound)
			return
		}
		payload.Version = version.Version
	}
	
	if err := s.siteService.WriteFile(slug, payload.Version, payload.File, payload.Content); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

func (s *Server) handleAPISiteWriteBatch(w http.ResponseWriter, r *http.Request, slug string) {
	var payload struct {
		Version string `json:"version"`
		Files   []struct {
			File    string `json:"file"`
			Content string `json:"content"`
		} `json:"files"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(payload.Files) == 0 {
		http.Error(w, "files array is required", http.StatusBadRequest)
		return
	}

	// Default to active version if not specified
	if payload.Version == "" {
		site, err := s.siteService.GetBySlug(slug)
		if err != nil {
			http.Error(w, "Site not found", http.StatusNotFound)
			return
		}
		versionRepo := db.NewSiteVersionRepository()
		version, err := versionRepo.GetActiveVersion(site.ID)
		if err != nil {
			http.Error(w, "No active version", http.StatusNotFound)
			return
		}
		payload.Version = version.Version
	}

	var written int
	for _, f := range payload.Files {
		if err := s.siteService.WriteFile(slug, payload.Version, f.File, f.Content); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(fmt.Sprintf(`{"success":false,"error":"write failed: %s","written":%d}`, err.Error(), written)))
			return
		}
		written++
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"success":true,"written":%d}`, written)))
}

func (s *Server) handleAPISiteUpload(w http.ResponseWriter, r *http.Request, slug string) {
	var payload struct {
		Path string `json:"path"`
		Data string `json:"data"` // base64-encoded
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success":false,"error":"invalid request body"}`))
		return
	}

	if payload.Path == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success":false,"error":"path is required"}`))
		return
	}
	if payload.Data == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success":false,"error":"data is required"}`))
		return
	}

	data, err := base64.StdEncoding.DecodeString(payload.Data)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success":false,"error":"invalid base64 data"}`))
		return
	}

	if err := s.siteService.UploadAsset(slug, payload.Path, data); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf(`{"success":false,"error":%q}`, err.Error())))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(fmt.Sprintf(`{"success":true,"path":%q,"size":%d}`, payload.Path, len(data))))
}

func (s *Server) handleAPISiteAssets(w http.ResponseWriter, r *http.Request, slug string) {
	path := strings.TrimPrefix(r.URL.Path, "/sites/"+slug+"/assets/")

	switch r.Method {
	case "GET":
		assets, err := s.siteService.ListAssets(slug)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf(`{"success":false,"error":%q}`, err.Error())))
			return
		}
		data, _ := json.Marshal(assets)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{"assets":%s}`, string(data))))

	case "DELETE":
		if path == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"success":false,"error":"asset path required"}`))
			return
		}
		if err := s.siteService.RemoveAsset(slug, path); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf(`{"success":false,"error":%q}`, err.Error())))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"success":true}`))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"success":false,"error":"method not allowed"}`))
	}
}

// authMiddleware validates Bearer token authentication
func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Public form submissions don't require auth
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/forms/") && strings.HasSuffix(r.URL.Path, "/submit") {
			next(w, r)
			return
		}

		// If no auth token configured, allow all requests
		if s.cfg.AuthToken == "" {
			next(w, r)
			return
		}
		
		// Check for Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: missing Authorization header", http.StatusUnauthorized)
			return
		}
		
		// Check Bearer token format
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			http.Error(w, "Unauthorized: invalid Authorization header format", http.StatusUnauthorized)
			return
		}
		
		token := authHeader[7:]
		if token != s.cfg.AuthToken {
			http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
			return
		}
		
		next(w, r)
	}
}

// resolveSite attempts to resolve a site slug from the request path or Host header.
// When accessed via a domain (e.g., test-site.intrane.fr/path), the site slug is
// looked up from the domain mapping, and the full path becomes the file path.
func (s *Server) resolveSite(path, host string) (siteSlug, filePath string, fromDomain bool) {
	// First, try path-based routing: /{slug}/{path}
	parts := strings.SplitN(path, "/", 2)
	siteSlug = parts[0]
	if len(parts) > 1 {
		filePath = parts[1]
	}

	// If path doesn't resolve to a known site, try Host header
	if _, err := s.siteService.GetBySlug(siteSlug); err != nil {
		h := extractHost(host)
		if domain, err := s.dnsService.GetDomainByDomain(h); err == nil && domain != nil {
			sites, _ := s.siteService.List()
			for _, site := range sites {
				if site.ID == domain.SiteID {
					return site.Slug, path, true
				}
			}
		}
	}

	return siteSlug, filePath, false
}

// extractHost strips the port from a Host header value.
func extractHost(host string) string {
	if idx := strings.LastIndex(host, ":"); idx >= 0 {
		return host[:idx]
	}
	return host
}

// isAssetExt returns true if the file path has a static asset extension.
func isAssetExt(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg",
		".webp", ".ico", ".woff", ".woff2", ".ttf", ".eot",
		".pdf", ".mp4", ".webm", ".json", ".xml":
		return true
	}
	return false
}

// tryServeAsset attempts to serve a file from the shared assets directory.
// Returns true if the asset was found and served.
func (s *Server) tryServeAsset(w http.ResponseWriter, r *http.Request, siteSlug, filePath string) bool {
	assetPath := filepath.Join(s.cfg.SitesDir, siteSlug, "assets", filePath)
	if _, err := os.Stat(assetPath); os.IsNotExist(err) {
		return false
	}

	// Detect content type from extension
	ctype := mime.TypeByExtension(filepath.Ext(filePath))
	if ctype == "" {
		ctype = "application/octet-stream"
	}

	data, err := os.ReadFile(assetPath)
	if err != nil {
		return false
	}

	w.Header().Set("Content-Type", ctype)
	w.Write(data)
	return true
}