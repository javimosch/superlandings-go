package db

import (
	"database/sql"
	"fmt"
	"time"
)

type SiteVersionRepository struct{}

func NewSiteVersionRepository() *SiteVersionRepository {
	return &SiteVersionRepository{}
}

func (r *SiteVersionRepository) Create(version *SiteVersion) error {
	query := `INSERT INTO site_versions (id, site_id, version, path, comment, author, is_active, orphaned, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`
	now := time.Now()
	version.CreatedAt = now
	_, err := DB.Exec(query, version.ID, version.SiteID, version.Version, version.Path,
		version.Comment, version.Author, version.IsActive, version.Orphaned, now)
	return err
}

func (r *SiteVersionRepository) GetActiveVersion(siteID string) (*SiteVersion, error) {
	query := `SELECT id, site_id, version, path, comment, author, is_active, orphaned, created_at
			  FROM site_versions WHERE site_id = ? AND is_active = 1`
	row := DB.QueryRow(query, siteID)
	return r.scanVersion(row)
}

func (r *SiteVersionRepository) GetBySiteAndVersion(siteID, version string) (*SiteVersion, error) {
	query := `SELECT id, site_id, version, path, comment, author, is_active, orphaned, created_at
			  FROM site_versions WHERE site_id = ? AND version = ?`
	row := DB.QueryRow(query, siteID, version)
	return r.scanVersion(row)
}

func (r *SiteVersionRepository) GetBySiteID(siteID string) ([]SiteVersion, error) {
	query := `SELECT id, site_id, version, path, comment, author, is_active, orphaned, created_at
			  FROM site_versions WHERE site_id = ? ORDER BY created_at DESC`
	return r.queryVersions(query, siteID)
}

func (r *SiteVersionRepository) ListVersions(siteID string) ([]SiteVersion, error) {
	return r.GetBySiteID(siteID)
}

func (r *SiteVersionRepository) SetActiveVersion(siteID string, version *SiteVersion) error {
	_, err := DB.Exec("UPDATE site_versions SET is_active = 0 WHERE site_id = ?", siteID)
	if err != nil {
		return fmt.Errorf("failed to deactivate versions: %w", err)
	}
	_, err = DB.Exec("UPDATE site_versions SET is_active = 1, orphaned = 0 WHERE id = ?", version.ID)
	if err != nil {
		return fmt.Errorf("failed to activate version: %w", err)
	}
	return nil
}

// MarkOrphanedAfter marks all versions after the given timestamp as orphaned.
// This is called after a rollback to flag dead-branch versions.
func (r *SiteVersionRepository) MarkOrphanedAfter(siteID string, afterTime time.Time) (int64, error) {
	result, err := DB.Exec(
		"UPDATE site_versions SET orphaned = 1 WHERE site_id = ? AND created_at > ?",
		siteID, afterTime,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetOrphanedVersions returns all orphaned (dead-branch) versions for a site.
func (r *SiteVersionRepository) GetOrphanedVersions(siteID string) ([]SiteVersion, error) {
	query := `SELECT id, site_id, version, path, comment, author, is_active, orphaned, created_at
			  FROM site_versions WHERE site_id = ? AND orphaned = 1 ORDER BY created_at DESC`
	return r.queryVersions(query, siteID)
}

// DeleteVersion removes a version row. Caller must delete filesystem dir first.
func (r *SiteVersionRepository) DeleteVersion(id string) error {
	_, err := DB.Exec("DELETE FROM site_versions WHERE id = ?", id)
	return err
}

// GetActiveLineage returns versions in the active branch (not orphaned, newest first).
func (r *SiteVersionRepository) GetActiveLineage(siteID string) ([]SiteVersion, error) {
	query := `SELECT id, site_id, version, path, comment, author, is_active, orphaned, created_at
			  FROM site_versions WHERE site_id = ? AND orphaned = 0 ORDER BY created_at DESC`
	return r.queryVersions(query, siteID)
}

func (r *SiteVersionRepository) scanVersion(row *sql.Row) (*SiteVersion, error) {
	var v SiteVersion
	err := row.Scan(&v.ID, &v.SiteID, &v.Version, &v.Path, &v.Comment, &v.Author, &v.IsActive, &v.Orphaned, &v.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("version not found")
		}
		return nil, fmt.Errorf("failed to scan version: %w", err)
	}
	return &v, nil
}

func (r *SiteVersionRepository) queryVersions(query string, args ...interface{}) ([]SiteVersion, error) {
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query versions: %w", err)
	}
	defer rows.Close()

	var versions []SiteVersion
	for rows.Next() {
		var v SiteVersion
		if err := rows.Scan(&v.ID, &v.SiteID, &v.Version, &v.Path, &v.Comment, &v.Author, &v.IsActive, &v.Orphaned, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, nil
}
