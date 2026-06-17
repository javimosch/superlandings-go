package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// LandingRepository handles landing database operations
type LandingRepository struct{}

func NewLandingRepository() *LandingRepository {
	return &LandingRepository{}
}

// Create creates a new landing
func (r *LandingRepository) Create(landing *Landing) error {
	query := `INSERT INTO landings (id, name, slug, type, organization_id, content, config, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	configJSON, err := json.Marshal(landing.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	now := time.Now()
	landing.CreatedAt = now
	landing.UpdatedAt = now

	_, err = DB.Exec(query, landing.ID, landing.Name, landing.Slug, landing.Type,
		landing.OrganizationID, landing.Content, string(configJSON), now, now)
	if err != nil {
		return fmt.Errorf("failed to create landing: %w", err)
	}

	// Insert files if virtual landing
	if landing.Type == "virtual" && len(landing.Files) > 0 {
		for _, file := range landing.Files {
			if err := r.createFile(landing.ID, file); err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
		}
	}

	// Insert domains
	for _, domain := range landing.Domains {
		if err := r.createDomain(landing.ID, domain); err != nil {
			return fmt.Errorf("failed to create domain: %w", err)
		}
	}

	return nil
}

// GetByID retrieves a landing by ID
func (r *LandingRepository) GetByID(id string) (*Landing, error) {
	query := `SELECT id, name, slug, type, organization_id, content, config, created_at, updated_at
			  FROM landings WHERE id = ?`

	row := DB.QueryRow(query, id)
	var landing Landing
	var configJSON string

	err := row.Scan(&landing.ID, &landing.Name, &landing.Slug, &landing.Type,
		&landing.OrganizationID, &landing.Content, &configJSON, &landing.CreatedAt, &landing.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("landing not found")
		}
		return nil, fmt.Errorf("failed to get landing: %w", err)
	}

	// Unmarshal config
	if configJSON != "" {
		if err := json.Unmarshal([]byte(configJSON), &landing.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	// Load files and domains
	files, err := r.getFiles(landing.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}
	landing.Files = files

	domains, err := r.getDomains(landing.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get domains: %w", err)
	}
	landing.Domains = domains

	return &landing, nil
}

// GetBySlug retrieves a landing by slug
func (r *LandingRepository) GetBySlug(slug string) (*Landing, error) {
	query := `SELECT id, name, slug, type, organization_id, content, config, created_at, updated_at
			  FROM landings WHERE slug = ?`

	row := DB.QueryRow(query, slug)
	var landing Landing
	var configJSON string

	err := row.Scan(&landing.ID, &landing.Name, &landing.Slug, &landing.Type,
		&landing.OrganizationID, &landing.Content, &configJSON, &landing.CreatedAt, &landing.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("landing not found")
		}
		return nil, fmt.Errorf("failed to get landing: %w", err)
	}

	// Unmarshal config
	if configJSON != "" {
		if err := json.Unmarshal([]byte(configJSON), &landing.Config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	}

	// Load files and domains
	files, err := r.getFiles(landing.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get files: %w", err)
	}
	landing.Files = files

	domains, err := r.getDomains(landing.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get domains: %w", err)
	}
	landing.Domains = domains

	return &landing, nil
}

// List retrieves all landings
func (r *LandingRepository) List() ([]Landing, error) {
	query := `SELECT id, name, slug, type, organization_id, content, config, created_at, updated_at
			  FROM landings ORDER BY created_at DESC`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list landings: %w", err)
	}
	defer rows.Close()

	var landings []Landing
	for rows.Next() {
		var landing Landing
		var configJSON string

		err := rows.Scan(&landing.ID, &landing.Name, &landing.Slug, &landing.Type,
			&landing.OrganizationID, &landing.Content, &configJSON, &landing.CreatedAt, &landing.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan landing: %w", err)
		}

		// Unmarshal config
		if configJSON != "" {
			if err := json.Unmarshal([]byte(configJSON), &landing.Config); err != nil {
				return nil, fmt.Errorf("failed to unmarshal config: %w", err)
			}
		}

		// Load files and domains
		files, err := r.getFiles(landing.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get files: %w", err)
		}
		landing.Files = files

		domains, err := r.getDomains(landing.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get domains: %w", err)
		}
		landing.Domains = domains

		landings = append(landings, landing)
	}

	return landings, nil
}

// Update updates a landing
func (r *LandingRepository) Update(landing *Landing) error {
	query := `UPDATE landings SET name = ?, slug = ?, type = ?, organization_id = ?, 
			  content = ?, config = ?, updated_at = ? WHERE id = ?`

	configJSON, err := json.Marshal(landing.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	landing.UpdatedAt = time.Now()

	_, err = DB.Exec(query, landing.Name, landing.Slug, landing.Type,
		landing.OrganizationID, landing.Content, string(configJSON), landing.UpdatedAt, landing.ID)
	if err != nil {
		return fmt.Errorf("failed to update landing: %w", err)
	}

	// Delete and recreate files
	if err := r.deleteFiles(landing.ID); err != nil {
		return fmt.Errorf("failed to delete files: %w", err)
	}
	for _, file := range landing.Files {
		if err := r.createFile(landing.ID, file); err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}
	}

	// Delete and recreate domains
	if err := r.deleteDomains(landing.ID); err != nil {
		return fmt.Errorf("failed to delete domains: %w", err)
	}
	for _, domain := range landing.Domains {
		if err := r.createDomain(landing.ID, domain); err != nil {
			return fmt.Errorf("failed to create domain: %w", err)
		}
	}

	return nil
}

// Delete deletes a landing
func (r *LandingRepository) Delete(id string) error {
	query := `DELETE FROM landings WHERE id = ?`
	_, err := DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete landing: %w", err)
	}
	return nil
}

// createFile creates a file for a landing
func (r *LandingRepository) createFile(landingID string, file File) error {
	query := `INSERT INTO landing_files (landing_id, path, content) VALUES (?, ?, ?)`
	_, err := DB.Exec(query, landingID, file.Path, file.Content)
	return err
}

// getFiles retrieves all files for a landing
func (r *LandingRepository) getFiles(landingID string) ([]File, error) {
	query := `SELECT path, content FROM landing_files WHERE landing_id = ?`
	rows, err := DB.Query(query, landingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var file File
		if err := rows.Scan(&file.Path, &file.Content); err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	return files, nil
}

// deleteFiles deletes all files for a landing
func (r *LandingRepository) deleteFiles(landingID string) error {
	query := `DELETE FROM landing_files WHERE landing_id = ?`
	_, err := DB.Exec(query, landingID)
	return err
}

// createDomain creates a domain for a landing
func (r *LandingRepository) createDomain(landingID string, domain Domain) error {
	query := `INSERT INTO landing_domains (landing_id, domain, traefik, cloudflare) VALUES (?, ?, ?, ?)`
	_, err := DB.Exec(query, landingID, domain.Domain, domain.Traefik, domain.Cloudflare)
	return err
}

// getDomains retrieves all domains for a landing
func (r *LandingRepository) getDomains(landingID string) ([]Domain, error) {
	query := `SELECT domain, traefik, cloudflare FROM landing_domains WHERE landing_id = ?`
	rows, err := DB.Query(query, landingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var domains []Domain
	for rows.Next() {
		var domain Domain
		if err := rows.Scan(&domain.Domain, &domain.Traefik, &domain.Cloudflare); err != nil {
			return nil, err
		}
		domains = append(domains, domain)
	}

	return domains, nil
}

// deleteDomains deletes all domains for a landing
func (r *LandingRepository) deleteDomains(landingID string) error {
	query := `DELETE FROM landing_domains WHERE landing_id = ?`
	_, err := DB.Exec(query, landingID)
	return err
}

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
