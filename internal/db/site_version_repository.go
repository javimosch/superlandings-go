package db

import (
	"database/sql"
	"fmt"
	"time"
)

// SiteVersionRepository handles site version database operations
type SiteVersionRepository struct{}

func NewSiteVersionRepository() *SiteVersionRepository {
	return &SiteVersionRepository{}
}

// Create creates a new site version
func (r *SiteVersionRepository) Create(version *SiteVersion) error {
	query := `INSERT INTO site_versions (id, site_id, version, path, comment, author, is_active, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	version.CreatedAt = now

	_, err := DB.Exec(query, version.ID, version.SiteID, version.Version, version.Path, version.Comment, version.Author, version.IsActive, now)
	return err
}

// GetActiveVersion retrieves the active version for a site
func (r *SiteVersionRepository) GetActiveVersion(siteID string) (*SiteVersion, error) {
	query := `SELECT id, site_id, version, path, comment, author, is_active, created_at FROM site_versions WHERE site_id = ? AND is_active = 1`

	row := DB.QueryRow(query, siteID)
	var version SiteVersion

	err := row.Scan(&version.ID, &version.SiteID, &version.Version, &version.Path, &version.Comment, &version.Author, &version.IsActive, &version.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active version found")
		}
		return nil, fmt.Errorf("failed to get active version: %w", err)
	}

	return &version, nil
}

// GetBySiteAndVersion retrieves a version by site ID and version string
func (r *SiteVersionRepository) GetBySiteAndVersion(siteID, version string) (*SiteVersion, error) {
	query := `SELECT id, site_id, version, path, comment, author, is_active, created_at FROM site_versions WHERE site_id = ? AND version = ?`

	row := DB.QueryRow(query, siteID, version)
	var v SiteVersion

	err := row.Scan(&v.ID, &v.SiteID, &v.Version, &v.Path, &v.Comment, &v.Author, &v.IsActive, &v.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("version not found")
		}
		return nil, fmt.Errorf("failed to get version: %w", err)
	}

	return &v, nil
}

// GetBySiteID retrieves all versions for a site
func (r *SiteVersionRepository) GetBySiteID(siteID string) ([]SiteVersion, error) {
	query := `SELECT id, site_id, version, path, comment, author, is_active, created_at FROM site_versions WHERE site_id = ? ORDER BY created_at DESC`

	rows, err := DB.Query(query, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query versions: %w", err)
	}
	defer rows.Close()

	var versions []SiteVersion
	for rows.Next() {
		var v SiteVersion
		if err := rows.Scan(&v.ID, &v.SiteID, &v.Version, &v.Path, &v.Comment, &v.Author, &v.IsActive, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, v)
	}

	return versions, nil
}

// ListVersions is an alias for GetBySiteID for compatibility
func (r *SiteVersionRepository) ListVersions(siteID string) ([]SiteVersion, error) {
	return r.GetBySiteID(siteID)
}

// SetActiveVersion sets a version as active for a site
func (r *SiteVersionRepository) SetActiveVersion(siteID string, version *SiteVersion) error {
	// Deactivate all versions for the site
	_, err := DB.Exec("UPDATE site_versions SET is_active = 0 WHERE site_id = ?", siteID)
	if err != nil {
		return fmt.Errorf("failed to deactivate versions: %w", err)
	}

	// Activate the specified version
	_, err = DB.Exec("UPDATE site_versions SET is_active = 1 WHERE id = ?", version.ID)
	if err != nil {
		return fmt.Errorf("failed to activate version: %w", err)
	}

	return nil
}