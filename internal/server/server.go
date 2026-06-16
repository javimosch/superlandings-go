package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/javimosch/superlandings-go/internal/services"
)

type Server struct {
	cfg          *config.Config
	landingService *services.LandingService
	siteService    *services.SiteService
}

func NewServer(cfg *config.Config) *Server {
	return &Server{
		cfg:          cfg,
		landingService: services.NewLandingService(cfg),
		siteService:    services.NewSiteService(cfg),
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
	mux.HandleFunc("/admin/", s.handleAdmin)

	// API routes must be registered before the catch-all / handler
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/status", s.authMiddleware(s.handleAPIStatus))
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

	if path == "" {
		s.handleRoot(w, r)
		return
	}

	// Try to serve as a site first (with dynamic blocks and sub-paths)
	// Extract site slug and file path
	parts := strings.SplitN(path, "/", 2)
	siteSlug := parts[0]
	filePath := ""
	if len(parts) > 1 {
		filePath = parts[1]
	}

	if content, err := s.siteService.GetActiveVersionContent(siteSlug, filePath); err == nil {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(content))
		return
	}

	// Fall back to landing
	content, contentType, err := s.landingService.GetLandingContent(path)
	if err != nil {
		http.Error(w, "Site or landing not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Write([]byte(content))
}

// handleRoot serves the root page
func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	// List all landings
	landings, err := s.landingService.List()
	if err != nil {
		http.Error(w, "Failed to list landings", http.StatusInternalServerError)
		return
	}

	// List all sites
	sites, err := s.siteService.List()
	if err != nil {
		http.Error(w, "Failed to list sites", http.StatusInternalServerError)
		return
	}

	html := "<html><head><title>SuperLandings</title></head><body>"
	html += "<h1>SuperLandings</h1>"
	
	// Sites section
	html += "<h2>Sites (with dynamic blocks):</h2>"
	html += "<ul>"
	for _, site := range sites {
		html += fmt.Sprintf("<li><a href=\"/%s\">%s</a></li>", site.Slug, site.Name)
	}
	if len(sites) == 0 {
		html += "<li>No sites found. Create one using: sl-cli site create</li>"
	}
	html += "</ul>"
	
	// Landings section
	html += "<h2>Landings:</h2>"
	html += "<ul>"
	for _, landing := range landings {
		html += fmt.Sprintf("<li><a href=\"/%s\">%s (%s)</a></li>", landing.Slug, landing.Name, landing.Type)
	}
	if len(landings) == 0 {
		html += "<li>No landings found. Create one using: sl-cli landing create</li>"
	}
	html += "</ul>"
	
	html += "</body></html>"

	w.Header().Set("Content-Type", "text/html")
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
	case "files":
		s.handleAdminAPIFiles(w, r)
	default:
		// Check if it's a file read path (sites/{slug}/files/{path})
		if strings.HasPrefix(action, "files/") {
			s.handleAdminAPIFileRead(w, r)
			return
		}
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
	
	if payload.Version == "" {
		http.Error(w, "version is required", http.StatusBadRequest)
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
	
	if err := s.siteService.WriteFile(slug, payload.Version, payload.File, payload.Content); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

// authMiddleware validates Bearer token authentication
func (s *Server) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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