package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

// Initialize creates the database connection and runs migrations
func Initialize(dbPath string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open database connection
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable WAL mode for better concurrency
	if _, err := DB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// Run migrations
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// runMigrations creates the database schema
func runMigrations() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS landings (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			type TEXT NOT NULL,
			organization_id TEXT,
			content TEXT,
			config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS landing_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			landing_id TEXT NOT NULL,
			path TEXT NOT NULL,
			content TEXT,
			FOREIGN KEY (landing_id) REFERENCES landings(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS landing_domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			landing_id TEXT NOT NULL,
			domain TEXT NOT NULL,
			traefik BOOLEAN DEFAULT 0,
			cloudflare BOOLEAN DEFAULT 0,
			FOREIGN KEY (landing_id) REFERENCES landings(id) ON DELETE CASCADE,
			UNIQUE(landing_id, domain)
		)`,
		`CREATE TABLE IF NOT EXISTS organizations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'viewer',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS sites (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			slug TEXT UNIQUE NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS site_versions (
			id TEXT PRIMARY KEY,
			site_id TEXT NOT NULL,
			version TEXT NOT NULL,
			path TEXT NOT NULL,
			comment TEXT,
			author TEXT,
			is_active BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
			UNIQUE(site_id, version)
		)`,
		`CREATE TABLE IF NOT EXISTS site_domains (
			id TEXT PRIMARY KEY,
			site_id TEXT NOT NULL,
			domain TEXT NOT NULL,
			ip TEXT,
			traefik BOOLEAN DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (site_id) REFERENCES sites(id) ON DELETE CASCADE,
			UNIQUE(site_id, domain)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_landings_slug ON landings(slug)`,
		`CREATE INDEX IF NOT EXISTS idx_landings_type ON landings(type)`,
		`CREATE INDEX IF NOT EXISTS idx_landings_org ON landings(organization_id)`,
		`CREATE INDEX IF NOT EXISTS idx_landing_files_landing_id ON landing_files(landing_id)`,
		`CREATE INDEX IF NOT EXISTS idx_landing_domains_landing_id ON landing_domains(landing_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sites_slug ON sites(slug)`,
		`CREATE INDEX IF NOT EXISTS idx_site_versions_site_id ON site_versions(site_id)`,
		`CREATE INDEX IF NOT EXISTS idx_site_versions_active ON site_versions(is_active)`,
		`CREATE INDEX IF NOT EXISTS idx_site_domains_site_id ON site_domains(site_id)`,
	}

	for _, migration := range migrations {
		if _, err := DB.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}