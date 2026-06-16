package db

import (
	"time"
)

// Landing represents a landing page
type Landing struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Slug           string    `json:"slug" db:"slug"`
	Type           string    `json:"type" db:"type"` // html, ejs, virtual, static, traefik-config
	OrganizationID string    `json:"organizationId,omitempty" db:"organization_id"`
	Content        string    `json:"content,omitempty" db:"content"` // for html type
	Files          []File    `json:"files,omitempty" db:"-"`         // for virtual type, stored separately
	Domains        []Domain  `json:"domains,omitempty" db:"-"`       // stored separately
	Config         Config    `json:"config,omitempty" db:"config"`
	CreatedAt      time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt      time.Time `json:"updatedAt" db:"updated_at"`
}

// File represents a file in a virtual landing
type File struct {
	Path    string `json:"path" db:"path"`
	Content string `json:"content" db:"content"`
}

// Domain represents a domain configuration
type Domain struct {
	Domain     string `json:"domain" db:"domain"`
	Traefik    bool   `json:"traefik" db:"traefik"`
	Cloudflare bool   `json:"cloudflare" db:"cloudflare"`
}

// Config represents landing configuration
type Config struct {
	SSLEnabled bool `json:"sslEnabled" db:"ssl_enabled"`
}

// Organization represents an organization
type Organization struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// User represents a user
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"` // admin, editor, viewer
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
}

// Site represents a static site with versions
type Site struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// SiteVersion represents a version of a site
type SiteVersion struct {
	ID        string    `json:"id" db:"id"`
	SiteID    string    `json:"siteId" db:"site_id"`
	Version   string    `json:"version" db:"version"` // "v1", "v2", etc.
	Path      string    `json:"path" db:"path"`       // FS path like "sites/foo/v1"
	Comment   string    `json:"comment" db:"comment"`
	Author    string    `json:"author" db:"author"`
	IsActive  bool      `json:"isActive" db:"is_active"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}