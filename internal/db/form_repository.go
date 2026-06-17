package db

import (
	"fmt"
	"time"
)

// FormRepository manages form submissions in SQLite
type FormRepository struct{}

// NewFormRepository creates a new form repository
func NewFormRepository() *FormRepository {
	return &FormRepository{}
}

// CreateSubmission stores a new form submission
func (r *FormRepository) CreateSubmission(id, siteID, formKey, formName, data string) error {
	_, err := DB.Exec(
		`INSERT INTO form_submissions (id, site_id, form_key, form_name, data, status, created_at)
		 VALUES (?, ?, ?, ?, ?, 'new', ?)`,
		id, siteID, formKey, formName, data, time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to create form submission: %w", err)
	}
	return nil
}

// ListSubmissions returns all submissions for a site and form key
func (r *FormRepository) ListSubmissions(siteID, formKey string, limit, offset int) ([]FormSubmission, error) {
	rows, err := DB.Query(
		`SELECT id, site_id, form_key, form_name, data, status, created_at
		 FROM form_submissions
		 WHERE site_id = ? AND form_key = ?
		 ORDER BY created_at DESC
		 LIMIT ? OFFSET ?`,
		siteID, formKey, limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query submissions: %w", err)
	}
	defer rows.Close()

	var submissions []FormSubmission
	for rows.Next() {
		var s FormSubmission
		var createdAt string
		if err := rows.Scan(&s.ID, &s.SiteID, &s.FormKey, &s.FormName, &s.Data, &s.Status, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan submission: %w", err)
		}
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		submissions = append(submissions, s)
	}
	return submissions, nil
}

// GetSubmission retrieves a single submission by ID
func (r *FormRepository) GetSubmission(id string) (*FormSubmission, error) {
	var s FormSubmission
	var createdAt string
	err := DB.QueryRow(
		`SELECT id, site_id, form_key, form_name, data, status, created_at
		 FROM form_submissions WHERE id = ?`, id,
	).Scan(&s.ID, &s.SiteID, &s.FormKey, &s.FormName, &s.Data, &s.Status, &createdAt)
	if err != nil {
		return nil, fmt.Errorf("submission not found: %w", err)
	}
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return &s, nil
}

// UpdateSubmissionStatus updates the status of a submission
func (r *FormRepository) UpdateSubmissionStatus(id, status string) error {
	_, err := DB.Exec(`UPDATE form_submissions SET status = ? WHERE id = ?`, status, id)
	if err != nil {
		return fmt.Errorf("failed to update submission: %w", err)
	}
	return nil
}

// DeleteSubmission removes a submission
func (r *FormRepository) DeleteSubmission(id string) error {
	_, err := DB.Exec(`DELETE FROM form_submissions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete submission: %w", err)
	}
	return nil
}

// CountSubmissions returns the total count for a site+form
func (r *FormRepository) CountSubmissions(siteID, formKey string) (int, error) {
	var count int
	err := DB.QueryRow(
		`SELECT COUNT(*) FROM form_submissions WHERE site_id = ? AND form_key = ?`,
		siteID, formKey,
	).Scan(&count)
	return count, err
}
