package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/yuin/goldmark"
)

// GetActiveVersionContent returns the processed content for the active version
func (s *SiteService) GetActiveVersionContent(siteSlug, filePath string) (string, error) {
	site, err := s.siteRepo.GetBySlug(siteSlug)
	if err != nil {
		return "", fmt.Errorf("site not found: %w", err)
	}

	// Get active version
	version, err := s.versionRepo.GetActiveVersion(site.ID)
	if err != nil {
		return "", fmt.Errorf("no active version: %w", err)
	}

	// Determine file path (default to index.html if not specified)
	if filePath == "" || filePath == "/" {
		filePath = "index.html"
	} else {
		// Remove leading slash
		filePath = strings.TrimPrefix(filePath, "/")

		// Try pages/ directory first
		pagesPath := filepath.Join("pages", filePath)
		if !strings.Contains(filePath, ".") {
			pagesPath += ".html"
		}

		// Try blog/ directory
		blogPath := filepath.Join("blog", filePath)
		if !strings.Contains(filePath, ".") {
			blogPath += ".md"
		}

		// Try each path in order: pages/, blog/, then root
		pathsToTry := []string{pagesPath, blogPath, filePath}
		versionDir := filepath.Join(s.cfg.SitesDir, site.Slug, version.Version)

		for _, path := range pathsToTry {
			fullPath := filepath.Join(versionDir, path)
			if content, err := os.ReadFile(fullPath); err == nil {
				return s.processContent(string(content), path, versionDir, siteSlug, site.ID)
			}
		}

		// If no extension, try adding .html
		if !strings.Contains(filePath, ".") {
			filePath += ".html"
		}
	}

	// Read the file
	indexPath := filepath.Join(s.cfg.SitesDir, site.Slug, version.Version, filePath)
	content, err := os.ReadFile(indexPath)
	if err != nil {
		return "", fmt.Errorf("failed to read %s: %w", filePath, err)
	}

	return s.processContent(string(content), filePath, filepath.Join(s.cfg.SitesDir, site.Slug, version.Version), siteSlug, site.ID)
}

// processContent processes file content with includes, templates, and auto-nav
func (s *SiteService) processContent(content, filePath, versionDir, siteSlug, siteID string) (string, error) {
	// Convert markdown to HTML if needed
	if strings.HasSuffix(filePath, ".md") {
		var buf bytes.Buffer
		md := goldmark.New()
		if err := md.Convert([]byte(content), &buf); err != nil {
			return "", fmt.Errorf("failed to convert markdown: %w", err)
		}
		content = buf.String()
	}

	// Process layout directive first
	processedContent := s.processLayout(content, versionDir)

	// If markdown and no layout specified, use default layout
	if strings.HasSuffix(filePath, ".md") && !strings.Contains(content, "{{>layout") {
		layoutPath := filepath.Join(versionDir, "layout.html")
		if layoutContent, err := os.ReadFile(layoutPath); err == nil {
			processedContent = strings.ReplaceAll(string(layoutContent), "{{.content}}", processedContent)
		}
	}

	// Process includes
	processedContent = s.processIncludes(processedContent, versionDir)

	// Check for data file (e.g., index.html.data.json)
	dataFilePath := filepath.Join(versionDir, filePath+".data.json")
	data, err := s.loadDataFile(dataFilePath)

	if err != nil {
		// No data file, create empty data map
		data = make(map[string]interface{})
	}

	// Auto-discover pages for navigation (if not in data)
	if _, ok := data["nav_pages"]; !ok {
		pages, err := s.DiscoverPages(siteSlug)
		if err == nil {
			data["nav_pages"] = pages
		}
	}

// Auto-discover blog posts for all pages
	if _, ok := data["blog_posts"]; !ok {
		posts, err := s.DiscoverBlogPosts(siteSlug)
		if err == nil {
			data["blog_posts"] = posts
		}
	}

	// Add root path variable
	data["root"] = "/" + siteSlug

	// Render with Go template
	renderedContent, err := s.renderTemplate(processedContent, data, versionDir)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	return renderedContent, nil
}

// WriteFile writes a file to a specific version
func (s *SiteService) WriteFile(siteSlug, version, filePath, content string) error {
	site, err := s.siteRepo.GetBySlug(siteSlug)
	if err != nil {
		return fmt.Errorf("site not found: %w", err)
	}

	// Create full path
	fullPath := filepath.Join(s.cfg.SitesDir, site.Slug, version, filePath)

	// Create directory if needed
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// GetVersionBySiteAndVersion gets a version by site ID and version string
func (s *SiteService) GetVersionBySiteAndVersion(siteID, version string) (*db.SiteVersion, error) {
	return s.versionRepo.GetBySiteAndVersion(siteID, version)
}

// Export exports site metadata to JSON
func (s *SiteService) Export(siteSlug string) (string, error) {
	site, err := s.siteRepo.GetBySlug(siteSlug)
	if err != nil {
		return "", fmt.Errorf("failed to get site: %w", err)
	}

	versions, err := s.versionRepo.ListVersions(site.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get versions: %w", err)
	}

	exportData := map[string]interface{}{
		"site":     site,
		"versions": versions,
	}

	jsonData, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal export: %w", err)
	}

	return string(jsonData), nil
}

// DiscoverPages discovers all pages in the pages/ directory
func (s *SiteService) DiscoverPages(siteSlug string) ([]map[string]interface{}, error) {
	site, err := s.siteRepo.GetBySlug(siteSlug)
	if err != nil {
		return nil, fmt.Errorf("site not found: %w", err)
	}

	// Get active version
	version, err := s.versionRepo.GetActiveVersion(site.ID)
	if err != nil {
		return nil, fmt.Errorf("no active version: %w", err)
	}

	// Check if pages/ directory exists
	pagesDir := filepath.Join(s.cfg.SitesDir, site.Slug, version.Version, "pages")
	if _, err := os.Stat(pagesDir); os.IsNotExist(err) {
		return []map[string]interface{}{}, nil
	}

	// Read all .html files in pages/
	entries, err := os.ReadDir(pagesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read pages directory: %w", err)
	}

	var pages []map[string]interface{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".html") {
			continue
		}

		// Extract slug (filename without .html)
		slug := strings.TrimSuffix(entry.Name(), ".html")

		// Try to load metadata from .data.json file
		dataFile := filepath.Join(pagesDir, entry.Name()+".data.json")
		metadata := make(map[string]interface{})
		if data, err := os.ReadFile(dataFile); err == nil {
			json.Unmarshal(data, &metadata)
		}

		// Set default title from slug if not in metadata
		if _, ok := metadata["title"]; !ok {
			metadata["title"] = strings.ReplaceAll(slug, "-", " ")
			metadata["title"] = strings.ToUpper(string(metadata["title"].(string))[0:1]) + metadata["title"].(string)[1:]
		}

		metadata["slug"] = slug
		pages = append(pages, metadata)
	}

	return pages, nil
}

// DiscoverBlogPosts discovers all blog posts in the blog/ directory
func (s *SiteService) DiscoverBlogPosts(siteSlug string) ([]map[string]interface{}, error) {
	site, err := s.siteRepo.GetBySlug(siteSlug)
	if err != nil {
		return nil, fmt.Errorf("site not found: %w", err)
	}

	// Get active version
	version, err := s.versionRepo.GetActiveVersion(site.ID)
	if err != nil {
		return nil, fmt.Errorf("no active version: %w", err)
	}

	// Check if blog/ directory exists
	blogDir := filepath.Join(s.cfg.SitesDir, site.Slug, version.Version, "blog")
	if _, err := os.Stat(blogDir); os.IsNotExist(err) {
		return []map[string]interface{}{}, nil
	}

	// Read all .md files in blog/
	entries, err := os.ReadDir(blogDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read blog directory: %w", err)
	}

	var posts []map[string]interface{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		// Extract slug (filename without .md)
		slug := strings.TrimSuffix(entry.Name(), ".md")

		// Try to load metadata from .data.json file
		dataFile := filepath.Join(blogDir, entry.Name()+".data.json")
		metadata := make(map[string]interface{})
		if data, err := os.ReadFile(dataFile); err == nil {
			json.Unmarshal(data, &metadata)
		}

		// Set default title from slug if not in metadata
		if _, ok := metadata["title"]; !ok {
			metadata["title"] = strings.ReplaceAll(slug, "-", " ")
			metadata["title"] = strings.ToUpper(string(metadata["title"].(string))[0:1]) + metadata["title"].(string)[1:]
		}

		metadata["slug"] = slug
		posts = append(posts, metadata)
	}

	return posts, nil
}