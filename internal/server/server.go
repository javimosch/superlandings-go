package server

import (
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
	// Initialize database
	if err := db.Initialize(s.cfg.DatabasePath); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Setup routes
	mux := http.NewServeMux()
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