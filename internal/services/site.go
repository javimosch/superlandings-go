package services

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
)

type SiteService struct {
	siteRepo    *db.SiteRepository
	versionRepo *db.SiteVersionRepository
	cfg         *config.Config
}

func NewSiteService(cfg *config.Config) *SiteService {
	return &SiteService{
		siteRepo:    db.NewSiteRepository(),
		versionRepo: db.NewSiteVersionRepository(),
		cfg:         cfg,
	}
}

// CreateSiteRequest represents the request to create a site
type CreateSiteRequest struct {
	Name string
	Slug string
}

// CreateVersionRequest represents the request to create a site version
type CreateVersionRequest struct {
	SiteID  string
	Version string
	Comment string
	Author  string
}

// Create creates a new site
func (s *SiteService) Create(req CreateSiteRequest) (*db.Site, error) {
	// Validate slug
	if req.Slug == "" {
		return nil, fmt.Errorf("slug is required")
	}
	if !isValidSlug(req.Slug) {
		return nil, fmt.Errorf("invalid slug format")
	}

	// Check if slug already exists
	if _, err := s.siteRepo.GetBySlug(req.Slug); err == nil {
		return nil, fmt.Errorf("slug already exists: %s", req.Slug)
	}

	// Create site directory
	siteDir := filepath.Join(s.cfg.SitesDir, req.Slug)
	if err := os.MkdirAll(siteDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create site directory: %w", err)
	}

	// Create site
	site := &db.Site{
		ID:   uuid.New().String(),
		Name: req.Name,
		Slug: req.Slug,
	}

	if err := s.siteRepo.Create(site); err != nil {
		return nil, fmt.Errorf("failed to create site: %w", err)
	}

	return site, nil
}

// GetBySlug retrieves a site by slug
func (s *SiteService) GetBySlug(slug string) (*db.Site, error) {
	return s.siteRepo.GetBySlug(slug)
}

// List retrieves all sites
func (s *SiteService) List() ([]db.Site, error) {
	return s.siteRepo.List()
}

// CreateVersion creates a new version of a site
func (s *SiteService) CreateVersion(siteID string, req CreateVersionRequest) (*db.SiteVersion, error) {
	// Get site to get slug
	site, err := s.siteRepo.GetBySlug(siteID)
	if err != nil {
		return nil, fmt.Errorf("site not found: %w", err)
	}

	// Validate version
	if req.Version == "" {
		return nil, fmt.Errorf("version is required")
	}

	// Create version directory
	versionDir := filepath.Join(s.cfg.SitesDir, site.Slug, req.Version)
	if err := os.MkdirAll(versionDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create version directory: %w", err)
	}

	// Check if version already exists
	versions, _ := s.versionRepo.ListVersions(site.ID)
	for _, v := range versions {
		if v.Version == req.Version {
			return nil, fmt.Errorf("version already exists: %s", req.Version)
		}
	}

	// Create version record
	version := &db.SiteVersion{
		ID:      uuid.New().String(),
		SiteID:  site.ID,
		Version: req.Version,
		Path:    filepath.Join("sites", site.Slug, req.Version),
		Comment: req.Comment,
		Author:  req.Author,
		IsActive: false,
	}

	if err := s.versionRepo.Create(version); err != nil {
		return nil, fmt.Errorf("failed to create version: %w", err)
	}

	// If this is the first version, make it active
	if len(versions) == 0 {
		if err := s.versionRepo.SetActiveVersion(site.ID, version.ID); err != nil {
			return nil, fmt.Errorf("failed to set active version: %w", err)
		}
		version.IsActive = true
	}

	return version, nil
}

// ListVersions retrieves all versions for a site
func (s *SiteService) ListVersions(siteSlug string) ([]db.SiteVersion, error) {
	site, err := s.siteRepo.GetBySlug(siteSlug)
	if err != nil {
		return nil, fmt.Errorf("site not found: %w", err)
	}

	return s.versionRepo.ListVersions(site.ID)
}

// SwitchVersion switches the active version for a site
func (s *SiteService) SwitchVersion(siteSlug, version string) error {
	site, err := s.siteRepo.GetBySlug(siteSlug)
	if err != nil {
		return fmt.Errorf("site not found: %w", err)
	}

	// Find the version
	versions, err := s.versionRepo.ListVersions(site.ID)
	if err != nil {
		return fmt.Errorf("failed to list versions: %w", err)
	}

	var targetVersion *db.SiteVersion
	for _, v := range versions {
		if v.Version == version {
			targetVersion = &v
			break
		}
	}

	if targetVersion == nil {
		return fmt.Errorf("version not found: %s", version)
	}

	// Set as active
	return s.versionRepo.SetActiveVersion(site.ID, targetVersion.ID)
}

// GetActiveVersionContent returns the processed content for the active version
func (s *SiteService) GetActiveVersionContent(siteSlug, filePath string) (string, error) {
	site, err := s.siteRepo.GetBySlug(siteSlug)
	if err != nil {
		return "", fmt.Errorf("site not found: %w", err)
	}

	// Get active version
	version, err := s.versionRepo.GetActiveVersion(site.ID)
	if err != nil {
		return "", fmt.Errorf("no active version: %w", err)
	}

	// Determine file path (default to index.html if not specified)
	if filePath == "" || filePath == "/" {
		filePath = "index.html"
	} else {
		// Remove leading slash and add .html if no extension
		filePath = strings.TrimPrefix(filePath, "/")
		if !strings.Contains(filePath, ".") {
			filePath += ".html"
		}
	}

	// Read the file
	indexPath := filepath.Join(s.cfg.SitesDir, site.Slug, version.Version, filePath)
	content, err := os.ReadFile(indexPath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	// Process includes first
	versionDir := filepath.Join(s.cfg.SitesDir, site.Slug, version.Version)
	processedContent := s.processIncludes(string(content), versionDir)

	// Check for data file (e.g., index.html.data.json)
	dataFilePath := indexPath + ".data.json"
	data, err := s.loadDataFile(dataFilePath)
	if err == nil {
		// Render with Go template
		renderedContent, err := s.renderTemplate(processedContent, data, versionDir)
		if err != nil {
			return "", fmt.Errorf("failed to render template: %w", err)
		}
		return renderedContent, nil
	}

	// No data file, return processed content as-is
	return processedContent, nil
}

// loadDataFile loads data from a .data.json file
func (s *SiteService) loadDataFile(dataPath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse data file: %w", err)
	}

	return result, nil
}

// renderTemplate renders content with Go's html/template
func (s *SiteService) renderTemplate(content string, data map[string]interface{}, baseDir string) (string, error) {
	// Create template
	tmpl, err := template.New("page").Parse(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Add custom functions
	tmpl = tmpl.Funcs(template.FuncMap{
		"include": func(path string) (string, error) {
			fullPath := filepath.Join(baseDir, path)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				return "", err
			}
			return string(content), nil
		},
	})

	// Render template
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// processIncludes processes {{>include "path"}} directives
func (s *SiteService) processIncludes(content string, basePath string) string {
	// Pattern to match {{>include "path"}}
	pattern := regexp.MustCompile(`{{>include "([^"]+)"}}`)

	return pattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the path
		matches := pattern.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match // Return original if no match
		}

		includePath := matches[1]
		fullPath := filepath.Join(basePath, includePath)

		// Read the included file
		if content, err := os.ReadFile(fullPath); err == nil {
			// Recursively process includes in the included file
			return s.processIncludes(string(content), basePath)
		}

		// If file not found, return original
		return match
	})
}

// WriteFile writes a file to a specific version
func (s *SiteService) WriteFile(siteSlug, version, filePath, content string) error {
	site, err := s.siteRepo.GetBySlug(siteSlug)
	if err != nil {
		return fmt.Errorf("site not found: %w", err)
	}

	// Create full path
	fullPath := filepath.Join(s.cfg.SitesDir, site.Slug, version, filePath)

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}