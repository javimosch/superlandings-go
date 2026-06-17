package server

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/javimosch/superlandings-go/internal/db"
)

// handleAPISiteForms handles form submissions (public) and admin listing (auth)
func (s *Server) handleAPISiteForms(w http.ResponseWriter, r *http.Request, slug string, parts []string) {
	if len(parts) < 2 {
		http.Error(w, "Invalid forms path", http.StatusBadRequest)
		return
	}

	formKey := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	// Get site
	sites, err := s.siteService.List()
	if err != nil {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}
	var siteID string
	for _, site := range sites {
		if site.Slug == slug {
			siteID = site.ID
			break
		}
	}
	if siteID == "" {
		http.Error(w, "Site not found", http.StatusNotFound)
		return
	}

	formRepo := db.NewFormRepository()

	switch {
	case action == "submit" && r.Method == "POST":
		s.handleFormSubmit(w, r, formRepo, siteID, formKey)
	case action == "submissions" && len(parts) == 2:
		s.handleFormSubmissionsList(w, r, formRepo, siteID, formKey)
	case action == "submissions" && len(parts) == 3 && parts[2] == "export":
		s.handleFormSubmissionExportCSV(w, r, formRepo, siteID, formKey)
	case action == "submissions" && len(parts) == 3:
		subID := parts[2]
		switch r.Method {
		case "GET":
			s.handleFormSubmissionGet(w, r, formRepo, subID)
		case "PATCH":
			s.handleFormSubmissionPatch(w, r, formRepo, subID)
		case "DELETE":
			s.handleFormSubmissionDelete(w, r, formRepo, subID)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	default:
		http.Error(w, "Invalid forms action", http.StatusBadRequest)
	}
}

// handleFormSubmit handles POST /api/sites/{slug}/forms/{key}/submit (public)
func (s *Server) handleFormSubmit(w http.ResponseWriter, r *http.Request, repo *db.FormRepository, siteID, formKey string) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success":false,"error":"Invalid JSON"}`))
		return
	}

	dataJSON, _ := json.Marshal(body)
	id := generateID()

	formName := formKey
	if n, ok := body["_form_name"].(string); ok && n != "" {
		formName = n
	}

	if err := repo.CreateSubmission(id, siteID, formKey, formName, string(dataJSON)); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success":false,"error":"Failed to save submission"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true,"message":"Submission received"}`))
}

// handleFormSubmissionsList returns all submissions (admin, auth required via middleware)
func (s *Server) handleFormSubmissionsList(w http.ResponseWriter, r *http.Request, repo *db.FormRepository, siteID, formKey string) {
	limit := 100
	offset := 0

	submissions, err := repo.ListSubmissions(siteID, formKey, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if submissions == nil {
		submissions = []db.FormSubmission{}
	}

	total, _ := repo.CountSubmissions(siteID, formKey)

	resp := map[string]interface{}{
		"submissions": submissions,
		"total":       total,
		"form_key":    formKey,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleFormSubmissionGet returns a single submission
func (s *Server) handleFormSubmissionGet(w http.ResponseWriter, r *http.Request, repo *db.FormRepository, id string) {
	sub, err := repo.GetSubmission(id)
	if err != nil {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

// handleFormSubmissionPatch updates status
func (s *Server) handleFormSubmissionPatch(w http.ResponseWriter, r *http.Request, repo *db.FormRepository, id string) {
	var body map[string]interface{}
	if json.NewDecoder(r.Body).Decode(&body) != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	status, _ := body["status"].(string)
	if status == "" {
		http.Error(w, "Missing status", http.StatusBadRequest)
		return
	}
	if err := repo.UpdateSubmissionStatus(id, status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

// handleFormSubmissionDelete removes a submission
func (s *Server) handleFormSubmissionDelete(w http.ResponseWriter, r *http.Request, repo *db.FormRepository, id string) {
	if err := repo.DeleteSubmission(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":true}`))
}

// handleFormSubmissionExportCSV exports submissions as CSV
func (s *Server) handleFormSubmissionExportCSV(w http.ResponseWriter, r *http.Request, repo *db.FormRepository, siteID, formKey string) {
	submissions, err := repo.ListSubmissions(siteID, formKey, 10000, 0)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-submissions.csv", formKey))

	writer := csv.NewWriter(w)
	writer.Write([]string{"ID", "Date", "Status", "Data"})

	for _, sub := range submissions {
		writer.Write([]string{
			sub.ID,
			sub.CreatedAt.Format(time.RFC3339),
			sub.Status,
			sub.Data,
		})
	}
	writer.Flush()
}
