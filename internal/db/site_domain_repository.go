package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// SiteDomainRepository handles site domain database operations
type SiteDomainRepository struct{}

func NewSiteDomainRepository() *SiteDomainRepository {
	return &SiteDomainRepository{}
}

// Create creates a new site domain
func (r *SiteDomainRepository) Create(domain *SiteDomain) error {
	query := `INSERT INTO site_domains (id, site_id, domain, ip, traefik, created_at)
			  VALUES (?, ?, ?, ?, ?, ?)`

	if domain.ID == "" {
		domain.ID = uuid.New().String()
	}

	domain.CreatedAt = time.Now()

	_, err := DB.Exec(query, domain.ID, domain.SiteID, domain.Domain, domain.IP, domain.Traefik, domain.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create site domain: %w", err)
	}

	return nil
}

// GetBySiteID returns all domains for a site
func (r *SiteDomainRepository) GetBySiteID(siteID string) ([]SiteDomain, error) {
	query := `SELECT id, site_id, domain, ip, traefik, created_at
			  FROM site_domains
			  WHERE site_id = ?
			  ORDER BY created_at DESC`

	rows, err := DB.Query(query, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to query site domains: %w", err)
	}
	defer rows.Close()

	var domains []SiteDomain
	for rows.Next() {
		var d SiteDomain
		var ip sql.NullString
		if err := rows.Scan(&d.ID, &d.SiteID, &d.Domain, &ip, &d.Traefik, &d.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan site domain: %w", err)
		}
		if ip.Valid {
			d.IP = ip.String
		}
		domains = append(domains, d)
	}

	return domains, nil
}

// GetByDomain returns a site domain by domain name
func (r *SiteDomainRepository) GetByDomain(domain string) (*SiteDomain, error) {
	query := `SELECT id, site_id, domain, ip, traefik, created_at
			  FROM site_domains
			  WHERE domain = ?`

	var d SiteDomain
	var ip sql.NullString
	err := DB.QueryRow(query, domain).Scan(&d.ID, &d.SiteID, &d.Domain, &ip, &d.Traefik, &d.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query site domain: %w", err)
	}
	if ip.Valid {
		d.IP = ip.String
	}

	return &d, nil
}

// Delete deletes a site domain
func (r *SiteDomainRepository) Delete(domainID string) error {
	query := `DELETE FROM site_domains WHERE id = ?`

	_, err := DB.Exec(query, domainID)
	if err != nil {
		return fmt.Errorf("failed to delete site domain: %w", err)
	}

	return nil
}

// DeleteBySiteID deletes all domains for a site
func (r *SiteDomainRepository) DeleteBySiteID(siteID string) error {
	query := `DELETE FROM site_domains WHERE site_id = ?`

	_, err := DB.Exec(query, siteID)
	if err != nil {
		return fmt.Errorf("failed to delete site domains: %w", err)
	}

	return nil
}