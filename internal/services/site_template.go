package services

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// loadDataFile loads data from a .data.json file
func (s *SiteService) loadDataFile(dataPath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(dataPath)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse data file: %w", err)
	}

	return result, nil
}

// renderTemplate renders content with Go's html/template
func (s *SiteService) renderTemplate(content string, data map[string]interface{}, baseDir string) (string, error) {
	// Extract site slug from data for asset resolution
	siteSlug := ""
	if root, ok := data["root"].(string); ok {
		siteSlug = strings.TrimPrefix(root, "/")
	}

	// Create template with functions registered BEFORE parsing
	tmpl := template.New("page").Funcs(template.FuncMap{
		"include": func(path string) (string, error) {
			fullPath := filepath.Join(baseDir, path)
			content, err := os.ReadFile(fullPath)
			if err != nil {
				return "", err
			}
			return string(content), nil
		},
		"asset": func(name string) string {
			if siteSlug == "" {
				return ""
			}
			assetsDir := filepath.Join(s.cfg.SitesDir, siteSlug, "assets")
			found := ""
			filepath.Walk(assetsDir, func(path string, info os.FileInfo, err error) error {
				if err != nil || found != "" {
					return nil
				}
				if !info.IsDir() && info.Name() == name {
					rel, _ := filepath.Rel(assetsDir, path)
					found = "/" + siteSlug + "/" + rel
				}
				return nil
			})
			return found
		},
	})

	tmpl, err := tmpl.Parse(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Render template
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// processLayout processes {{>layout "path"}} directives
func (s *SiteService) processLayout(content string, basePath string) string {
	// Pattern to match {{>layout "path"}}
	pattern := regexp.MustCompile(`{{>layout "([^"]+)"}}`)

	matches := pattern.FindStringSubmatch(content)
	if len(matches) < 2 {
		return content // No layout directive
	}

	layoutPath := matches[1]
	layoutFullPath := filepath.Join(basePath, layoutPath)

	// Read the layout file
	layoutContent, err := os.ReadFile(layoutFullPath)
	if err != nil {
		return content // Layout file not found, return original
	}

	// Remove the layout directive from content
	pageContent := pattern.ReplaceAllString(content, "")

	// Replace {{.content}} in layout with page content
	layoutWithContent := strings.ReplaceAll(string(layoutContent), "{{.content}}", pageContent)

	return layoutWithContent
}

// processIncludes processes {{>include "path"}} directives
func (s *SiteService) processIncludes(content string, basePath string) string {
	// Pattern to match {{>include "path"}}
	pattern := regexp.MustCompile(`{{>include "([^"]+)"}}`)

	return pattern.ReplaceAllStringFunc(content, func(match string) string {
		// Extract the path
		matches := pattern.FindStringSubmatch(match)
		if len(matches) < 2 {
			return match // Return original if no match
		}

		includePath := matches[1]
		fullPath := filepath.Join(basePath, includePath)

		// Read the included file
		if content, err := os.ReadFile(fullPath); err == nil {
			// Recursively process includes in the included file
			return s.processIncludes(string(content), basePath)
		}

		// If file not found, return original
		return match
	})
}