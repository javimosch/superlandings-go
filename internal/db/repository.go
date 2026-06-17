package db

import (
	"database/sql"
	"fmt"
	"time"
)

// SiteRepository handles site database operations
type SiteRepository struct{}

func NewSiteRepository() *SiteRepository {
	return &SiteRepository{}
}

// Create creates a new site
func (r *SiteRepository) Create(site *Site) error {
	query := `INSERT INTO sites (id, name, slug, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`

	now := time.Now()
	site.CreatedAt = now
	site.UpdatedAt = now

	_, err := DB.Exec(query, site.ID, site.Name, site.Slug, now, now)
	return err
}

// GetBySlug retrieves a site by slug
func (r *SiteRepository) GetBySlug(slug string) (*Site, error) {
	query := `SELECT id, name, slug, created_at, updated_at FROM sites WHERE slug = ?`

	row := DB.QueryRow(query, slug)
	var site Site

	err := row.Scan(&site.ID, &site.Name, &site.Slug, &site.CreatedAt, &site.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("site not found")
		}
		return nil, fmt.Errorf("failed to get site: %w", err)
	}

	return &site, nil
}

// List retrieves all sites
func (r *SiteRepository) List() ([]Site, error) {
	query := `SELECT id, name, slug, created_at, updated_at FROM sites ORDER BY created_at DESC`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list sites: %w", err)
	}
	defer rows.Close()

	var sites []Site
	for rows.Next() {
		var site Site
		err := rows.Scan(&site.ID, &site.Name, &site.Slug, &site.CreatedAt, &site.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan site: %w", err)
		}
		sites = append(sites, site)
	}

	return sites, nil
}
// DeleteSite removes a site and all related data (cascaded via FK)
func (r *SiteRepository) DeleteSite(slug string) error {
	_, err := DB.Exec(`DELETE FROM sites WHERE slug = ?`, slug)
	return err
}
