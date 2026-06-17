package services

import (
	"fmt"
	"os"
	"path/filepath"

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
		if err := s.versionRepo.SetActiveVersion(site.ID, version); err != nil {
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
	return s.versionRepo.SetActiveVersion(site.ID, targetVersion)
}
func isValidSlug(slug string) bool {
	if len(slug) == 0 {
		return false
	}
	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}
