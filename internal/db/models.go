package db

import (
	"time"
)

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
	Version   string    `json:"version" db:"version"`
	Path      string    `json:"path" db:"path"`
	Comment   string    `json:"comment" db:"comment"`
	Author    string    `json:"author" db:"author"`
	IsActive  bool      `json:"isActive" db:"is_active"`
	Orphaned  bool      `json:"orphaned" db:"orphaned"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// SiteDomain represents a domain configuration for a site
type SiteDomain struct {
	ID        string    `json:"id" db:"id"`
	SiteID    string    `json:"siteId" db:"site_id"`
	Domain    string    `json:"domain" db:"domain"`
	IP        string    `json:"ip" db:"ip"`
	Traefik   bool      `json:"traefik" db:"traefik"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// SiteUser represents a user's access to a site
type SiteUser struct {
	ID        string    `json:"id" db:"id"`
	SiteID    string    `json:"siteId" db:"site_id"`
	UserID    string    `json:"userId" db:"user_id"`
	Role      string    `json:"role" db:"role"` // editor, viewer
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// SiteAdminToken represents an admin access token for a site
type SiteAdminToken struct {
	ID        string     `json:"id" db:"id"`
	SiteID    string     `json:"siteId" db:"site_id"`
	Token     string     `json:"token" db:"token"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty" db:"expires_at"`
	IsActive  bool       `json:"isActive" db:"is_active"`
}

// FormSubmission represents a submitted form entry
type FormSubmission struct {
	ID        string    `json:"id"`
	SiteID    string    `json:"siteId"`
	FormKey   string    `json:"formKey"`
	FormName  string    `json:"formName"`
	Data      string    `json:"data"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}