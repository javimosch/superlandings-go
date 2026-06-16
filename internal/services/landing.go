package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
)

type LandingService struct {
	repo *db.LandingRepository
	cfg  *config.Config
}

func NewLandingService(cfg *config.Config) *LandingService {
	return &LandingService{
		repo: db.NewLandingRepository(),
		cfg:  cfg,
	}
}

// CreateLandingRequest represents the request to create a landing
type CreateLandingRequest struct {
	Name           string
	Slug           string
	Type           string
	OrganizationID string
	Content        string
	Files          []db.File
	Domains        []db.Domain
}

// UpdateLandingRequest represents the request to update a landing
type UpdateLandingRequest struct {
	Name      string
	Slug      string
	Type      string
	Content   string
	Files     []db.File
	Domains   []db.Domain
}

// Create creates a new landing
func (s *LandingService) Create(req CreateLandingRequest) (*db.Landing, error) {
	// Validate type
	validTypes := []string{"html", "ejs", "virtual", "static", "traefik-config"}
	if !contains(validTypes, req.Type) {
		return nil, fmt.Errorf("invalid type: %s (valid types: %s)", req.Type, strings.Join(validTypes, ", "))
	}

	// Validate slug
	if req.Slug == "" {
		return nil, fmt.Errorf("slug is required")
	}
	if !isValidSlug(req.Slug) {
		return nil, fmt.Errorf("invalid slug format")
	}

	// Check if slug already exists
	if _, err := s.repo.GetBySlug(req.Slug); err == nil {
		return nil, fmt.Errorf("slug already exists: %s", req.Slug)
	}

	// Create landing directory
	landingDir := filepath.Join(s.cfg.LandingsDir, req.Slug)
	if err := os.MkdirAll(landingDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create landing directory: %w", err)
	}

	// Create landing
	landing := &db.Landing{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Slug:           req.Slug,
		Type:           req.Type,
		OrganizationID: req.OrganizationID,
		Content:        req.Content,
		Files:          req.Files,
		Domains:        req.Domains,
		Config:         db.Config{SSLEnabled: false},
	}

	// Save to database
	if err := s.repo.Create(landing); err != nil {
		return nil, fmt.Errorf("failed to create landing: %w", err)
	}

	// Create files on disk
	if req.Type == "html" || req.Type == "ejs" {
		if req.Content != "" {
			indexPath := filepath.Join(landingDir, "index.html")
			if req.Type == "ejs" {
				indexPath = filepath.Join(landingDir, "index.ejs")
			}
			if err := os.WriteFile(indexPath, []byte(req.Content), 0644); err != nil {
				return nil, fmt.Errorf("failed to write index file: %w", err)
			}
		}
	} else if req.Type == "virtual" {
		for _, file := range req.Files {
			filePath := filepath.Join(landingDir, file.Path)
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create file directory: %w", err)
			}
			if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
				return nil, fmt.Errorf("failed to write file: %w", err)
			}
		}
	}

	return landing, nil
}

// GetByID retrieves a landing by ID
func (s *LandingService) GetByID(id string) (*db.Landing, error) {
	return s.repo.GetByID(id)
}

// GetBySlug retrieves a landing by slug
func (s *LandingService) GetBySlug(slug string) (*db.Landing, error) {
	return s.repo.GetBySlug(slug)
}

// List retrieves all landings
func (s *LandingService) List() ([]db.Landing, error) {
	return s.repo.List()
}

// Update updates a landing
func (s *LandingService) Update(id string, req UpdateLandingRequest) (*db.Landing, error) {
	// Get existing landing
	landing, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("landing not found: %w", err)
	}

	// Validate type if changing
	if req.Type != "" {
		validTypes := []string{"html", "ejs", "virtual", "static", "traefik-config"}
		if !contains(validTypes, req.Type) {
			return nil, fmt.Errorf("invalid type: %s", req.Type)
		}
		landing.Type = req.Type
	}

	// Validate slug if changing
	if req.Slug != "" && req.Slug != landing.Slug {
		if !isValidSlug(req.Slug) {
			return nil, fmt.Errorf("invalid slug format")
		}
		// Check if new slug already exists
		if _, err := s.repo.GetBySlug(req.Slug); err == nil {
			return nil, fmt.Errorf("slug already exists: %s", req.Slug)
		}
		landing.Slug = req.Slug
	}

	// Update fields
	if req.Name != "" {
		landing.Name = req.Name
	}
	if req.Content != "" {
		landing.Content = req.Content
	}
	if req.Files != nil {
		landing.Files = req.Files
	}
	if req.Domains != nil {
		landing.Domains = req.Domains
	}

	// Save to database
	if err := s.repo.Update(landing); err != nil {
		return nil, fmt.Errorf("failed to update landing: %w", err)
	}

	// Update files on disk
	landingDir := filepath.Join(s.cfg.LandingsDir, landing.Slug)
	if landing.Type == "html" || landing.Type == "ejs" {
		if landing.Content != "" {
			indexPath := filepath.Join(landingDir, "index.html")
			if landing.Type == "ejs" {
				indexPath = filepath.Join(landingDir, "index.ejs")
			}
			if err := os.WriteFile(indexPath, []byte(landing.Content), 0644); err != nil {
				return nil, fmt.Errorf("failed to write index file: %w", err)
			}
		}
	} else if landing.Type == "virtual" {
		// Recreate all files
		for _, file := range landing.Files {
			filePath := filepath.Join(landingDir, file.Path)
			if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create file directory: %w", err)
			}
			if err := os.WriteFile(filePath, []byte(file.Content), 0644); err != nil {
				return nil, fmt.Errorf("failed to write file: %w", err)
			}
		}
	}

	return landing, nil
}

// Delete deletes a landing
func (s *LandingService) Delete(id string) error {
	// Get landing to get slug
	landing, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("landing not found: %w", err)
	}

	// Delete from database
	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete landing: %w", err)
	}

	// Delete files on disk
	landingDir := filepath.Join(s.cfg.LandingsDir, landing.Slug)
	if err := os.RemoveAll(landingDir); err != nil {
		return fmt.Errorf("failed to delete landing directory: %w", err)
	}

	return nil
}

// GetLandingContent returns the content for serving a landing
func (s *LandingService) GetLandingContent(slug string) (string, string, error) {
	landing, err := s.GetBySlug(slug)
	if err != nil {
		return "", "", fmt.Errorf("landing not found: %w", err)
	}

	landingDir := filepath.Join(s.cfg.LandingsDir, landing.Slug)

	switch landing.Type {
	case "html":
		content, err := os.ReadFile(filepath.Join(landingDir, "index.html"))
		if err != nil {
			return "", "", fmt.Errorf("failed to read index.html: %w", err)
		}
		return string(content), "text/html", nil
	case "ejs":
		content, err := os.ReadFile(filepath.Join(landingDir, "index.ejs"))
		if err != nil {
			return "", "", fmt.Errorf("failed to read index.ejs: %w", err)
		}
		return string(content), "text/html", nil
	case "virtual":
		// For virtual, return index.html if exists, otherwise first file
		if indexPath := filepath.Join(landingDir, "index.html"); fileExists(indexPath) {
			content, err := os.ReadFile(indexPath)
			if err != nil {
				return "", "", fmt.Errorf("failed to read index.html: %w", err)
			}
			return string(content), "text/html", nil
		}
		if len(landing.Files) > 0 {
			return landing.Files[0].Content, "text/html", nil
		}
		return "", "", fmt.Errorf("no content found for virtual landing")
	default:
		return "", "", fmt.Errorf("unsupported landing type: %s", landing.Type)
	}
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func isValidSlug(slug string) bool {
	if len(slug) == 0 {
		return false
	}
	// Allow alphanumeric, hyphens, and underscores
	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}