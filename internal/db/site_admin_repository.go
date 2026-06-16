package db

import (
	"database/sql"
	"fmt"
	"time"
)

// SiteAdminRepository handles site admin token operations
type SiteAdminRepository struct{}

func NewSiteAdminRepository() *SiteAdminRepository {
	return &SiteAdminRepository{}
}

// CreateAdminToken creates a new admin token for a site
func (r *SiteAdminRepository) CreateAdminToken(siteID, token string, expiresAt *time.Time) error {
	query := `INSERT INTO site_admin_tokens (id, site_id, token, created_at, expires_at, is_active)
			  VALUES (?, ?, ?, ?, ?, ?)`

	now := time.Now()
	id := fmt.Sprintf("admin-%s-%d", siteID, now.Unix())

	_, err := DB.Exec(query, id, siteID, token, now, expiresAt, true)
	if err != nil {
		return fmt.Errorf("failed to create admin token: %w", err)
	}

	return nil
}

// GetActiveTokenBySite retrieves the active admin token for a site
func (r *SiteAdminRepository) GetActiveTokenBySite(siteID string) (*SiteAdminToken, error) {
	query := `SELECT id, site_id, token, created_at, expires_at, is_active
			  FROM site_admin_tokens WHERE site_id = ? AND is_active = 1
			  ORDER BY created_at DESC LIMIT 1`

	row := DB.QueryRow(query, siteID)
	var token SiteAdminToken

	var expiresAt sql.NullTime
	err := row.Scan(&token.ID, &token.SiteID, &token.Token, &token.CreatedAt, &expiresAt, &token.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no active token found")
		}
		return nil, fmt.Errorf("failed to get admin token: %w", err)
	}

	if expiresAt.Valid {
		token.ExpiresAt = &expiresAt.Time
	}

	return &token, nil
}

// GetTokenByValue retrieves a token by its value
func (r *SiteAdminRepository) GetTokenByValue(token string) (*SiteAdminToken, error) {
	query := `SELECT id, site_id, token, created_at, expires_at, is_active
			  FROM site_admin_tokens WHERE token = ?`

	row := DB.QueryRow(query, token)
	var adminToken SiteAdminToken

	var expiresAt sql.NullTime
	err := row.Scan(&adminToken.ID, &adminToken.SiteID, &adminToken.Token, &adminToken.CreatedAt, &expiresAt, &adminToken.IsActive)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	if expiresAt.Valid {
		adminToken.ExpiresAt = &expiresAt.Time
	}

	return &adminToken, nil
}

// RotateToken creates a new token and deactivates old ones
func (r *SiteAdminRepository) RotateToken(siteID, newToken string, expiresAt *time.Time) error {
	// Deactivate all existing tokens
	deactivateQuery := `UPDATE site_admin_tokens SET is_active = 0 WHERE site_id = ?`
	if _, err := DB.Exec(deactivateQuery, siteID); err != nil {
		return fmt.Errorf("failed to deactivate old tokens: %w", err)
	}

	// Create new token
	return r.CreateAdminToken(siteID, newToken, expiresAt)
}

// RevokeAllTokens deactivates all tokens for a site
func (r *SiteAdminRepository) RevokeAllTokens(siteID string) error {
	query := `UPDATE site_admin_tokens SET is_active = 0 WHERE site_id = ?`
	_, err := DB.Exec(query, siteID)
	if err != nil {
		return fmt.Errorf("failed to revoke tokens: %w", err)
	}
	return nil
}