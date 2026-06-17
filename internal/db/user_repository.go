package db

import (
	"database/sql"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserRepository handles user database operations
type UserRepository struct{}

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// Create creates a new user
func (r *UserRepository) Create(user *User, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = string(hash)
	query := `INSERT INTO users (id, email, password_hash, role, created_at, updated_at)
			  VALUES (?, ?, ?, ?, ?, ?)`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err = DB.Exec(query, user.ID, user.Email, user.PasswordHash, user.Role, now, now)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(email string) (*User, error) {
	query := `SELECT id, email, password_hash, role, created_at, updated_at
			  FROM users WHERE email = ?`

	row := DB.QueryRow(query, email)
	var user User

	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id string) (*User, error) {
	query := `SELECT id, email, password_hash, role, created_at, updated_at FROM users WHERE id = ?`
	row := DB.QueryRow(query, id)
	var user User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return &user, nil
}

// List retrieves all users
func (r *UserRepository) List() ([]User, error) {
	query := `SELECT id, email, password_hash, role, created_at, updated_at
			  FROM users ORDER BY created_at DESC`

	rows, err := DB.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdatePassword updates a user's password
func (r *UserRepository) UpdatePassword(email string, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `UPDATE users SET password_hash = ?, updated_at = ? WHERE email = ?`
	now := time.Now()

	_, err = DB.Exec(query, string(hash), now, email)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// VerifyPassword verifies a user's password
func (r *UserRepository) VerifyPassword(email string, password string) (bool, error) {
	user, err := r.GetByEmail(email)
	if err != nil {
		return false, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return false, nil
	}

	return true, nil
}

// GrantSiteAccess grants a user access to a site
func (r *UserRepository) GrantSiteAccess(siteID, userID, role string) error {
	query := `INSERT INTO site_users (id, site_id, user_id, role, created_at)
			  VALUES (?, ?, ?, ?, ?)`

	now := time.Now()
	id := fmt.Sprintf("%s-%s", siteID, userID)

	_, err := DB.Exec(query, id, siteID, userID, role, now)
	if err != nil {
		return fmt.Errorf("failed to grant site access: %w", err)
	}

	return nil
}

// GetSiteUsers retrieves all users with access to a site
func (r *UserRepository) GetSiteUsers(siteID string) ([]SiteUser, error) {
	query := `SELECT id, site_id, user_id, role, created_at
			  FROM site_users WHERE site_id = ?`

	rows, err := DB.Query(query, siteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get site users: %w", err)
	}
	defer rows.Close()

	var siteUsers []SiteUser
	for rows.Next() {
		var su SiteUser
		err := rows.Scan(&su.ID, &su.SiteID, &su.UserID, &su.Role, &su.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan site user: %w", err)
		}
		siteUsers = append(siteUsers, su)
	}

	return siteUsers, nil
}

// RevokeSiteAccess revokes a user's access to a site
func (r *UserRepository) RevokeSiteAccess(siteID, userID string) error {
	query := `DELETE FROM site_users WHERE site_id = ? AND user_id = ?`
	_, err := DB.Exec(query, siteID, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke site access: %w", err)
	}
	return nil
}

// UserSiteInfo holds a site + role for dashboard listing
type UserSiteInfo struct {
	SiteID   string `json:"siteId"`
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

// GetUserSites returns all sites a user has access to
func (r *UserRepository) GetUserSites(email string) ([]UserSiteInfo, error) {
	query := `
		SELECT s.id, s.slug, s.name, su.role
		FROM site_users su
		JOIN sites s ON s.id = su.site_id
		JOIN users u ON u.id = su.user_id
		WHERE u.email = ?
		ORDER BY s.name
	`
	rows, err := DB.Query(query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sites: %w", err)
	}
	defer rows.Close()

	var result []UserSiteInfo
	for rows.Next() {
		var s UserSiteInfo
		if err := rows.Scan(&s.SiteID, &s.Slug, &s.Name, &s.Role); err != nil {
			return nil, fmt.Errorf("failed to scan user site: %w", err)
		}
		result = append(result, s)
	}
	return result, nil
}