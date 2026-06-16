package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/javimosch/superlandings-go/internal/config"
)

type SyncService struct {
	cfg         *config.Config
	siteService *SiteService
}

func NewSyncService(cfg *config.Config) *SyncService {
	return &SyncService{
		cfg:         cfg,
		siteService: NewSiteService(cfg),
	}
}

// Export exports site metadata to JSON
func (s *SyncService) Export(siteSlug string) (string, error) {
	site, err := s.siteService.GetBySlug(siteSlug)
	if err != nil {
		return "", fmt.Errorf("failed to get site: %w", err)
	}

	versions, err := s.siteService.ListVersions(siteSlug)
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

// Import imports site metadata from JSON
func (s *SyncService) Import(jsonData string) error {
	var importData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &importData); err != nil {
		return fmt.Errorf("failed to unmarshal import: %w", err)
	}

	// Import site
	siteData, ok := importData["site"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid site data in import")
	}

	siteID := siteData["id"].(string)
	name := siteData["name"].(string)
	slug := siteData["slug"].(string)

	// Check if site exists, create if not
	existingSite, err := s.siteService.GetBySlug(slug)
	if err != nil {
		// Create new site
		req := CreateSiteRequest{
			Name: name,
			Slug: slug,
		}
		site, err := s.siteService.Create(req)
		if err != nil {
			return fmt.Errorf("failed to create site: %w", err)
		}
		siteID = site.ID
	} else {
		siteID = existingSite.ID
	}

	// Import versions
	versionsData, ok := importData["versions"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid versions data in import")
	}

	for _, v := range versionsData {
		versionData := v.(map[string]interface{})
		version := versionData["version"].(string)
		comment := ""
		if c, ok := versionData["comment"].(string); ok {
			comment = c
		}
		author := ""
		if a, ok := versionData["author"].(string); ok {
			author = a
		}

		// Check if version exists
		_, err := s.siteService.GetVersionBySiteAndVersion(siteID, version)
		if err != nil {
			// Create new version
			req := CreateVersionRequest{
				SiteID:  siteID,
				Version: version,
				Comment: comment,
				Author:  author,
			}
			if _, err := s.siteService.CreateVersion(siteID, req); err != nil {
				return fmt.Errorf("failed to create version %s: %w", version, err)
			}
		}
	}

	return nil
}

// SyncTarget represents a remote sync target
type SyncTarget struct {
	Host string
	User string
	Port int
	Key  string // SSH key path
}

// Sync syncs a site to a remote target
func (s *SyncService) Sync(siteSlug string, target SyncTarget) error {
	// Export site metadata
	exportData, err := s.siteService.Export(siteSlug)
	if err != nil {
		return fmt.Errorf("failed to export site: %w", err)
	}

	// Save export to temp file
	tempFile := "/tmp/site-export.json"
	if err := os.WriteFile(tempFile, []byte(exportData), 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}
	defer os.Remove(tempFile)

	// Sync site directory via rsync
	sitePath := filepath.Join(s.cfg.SitesDir, siteSlug)
	portFlag := ""
	if target.Port != 22 {
		portFlag = fmt.Sprintf("-p %d", target.Port)
	}
	remotePath := fmt.Sprintf("%s@%s:%s/.superlandings/sites/%s",
		target.User, target.Host, portFlag, siteSlug)

	rsyncArgs := []string{"-avz"}
	if target.Key != "" {
		rsyncArgs = append(rsyncArgs, "-e", fmt.Sprintf("ssh -i %s -o IdentitiesOnly=yes", target.Key))
	}
	rsyncArgs = append(rsyncArgs, sitePath+"/", remotePath+"/")

	rsyncCmd := exec.Command("rsync", rsyncArgs...)
	if output, err := rsyncCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to rsync site files: %w, output: %s", err, string(output))
	}

	// Copy export file to remote
	scpArgs := []string{}
	if target.Key != "" {
		scpArgs = append(scpArgs, "-i", target.Key, "-o", "IdentitiesOnly=yes")
	}
	scpArgs = append(scpArgs, tempFile, fmt.Sprintf("%s@%s:/tmp/site-import.json", target.User, target.Host))

	scpCmd := exec.Command("scp", scpArgs...)
	if output, err := scpCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to copy export file: %w, output: %s", err, string(output))
	}

	// Import metadata on remote
	sshArgs := []string{}
	if target.Key != "" {
		sshArgs = append(sshArgs, "-i", target.Key, "-o", "IdentitiesOnly=yes")
	}
	sshArgs = append(sshArgs, fmt.Sprintf("%s@%s", target.User, target.Host), "sl-cli site import --input /tmp/site-import.json")

	sshCmd := exec.Command("ssh", sshArgs...)
	if output, err := sshCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to import on remote: %w, output: %s", err, string(output))
	}

	// Restart daemon on remote to pick up changes
	restartArgs := []string{}
	if target.Key != "" {
		restartArgs = append(restartArgs, "-i", target.Key, "-o", "IdentitiesOnly=yes")
	}
	restartArgs = append(restartArgs, fmt.Sprintf("%s@%s", target.User, target.Host), "pkill -f 'sl-cli backend' && sl-cli backend start --daemon --port 3100")

	restartCmd := exec.Command("ssh", restartArgs...)
	if output, err := restartCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to restart remote daemon: %w, output: %s", err, string(output))
	}

	return nil
}

// ProxySetup represents proxy configuration
type ProxySetup struct {
	SiteSlug    string
	Domain      string
	InternalURL string
}

// SetupProxy configures hotify-cli reverse proxy for a site
func (s *SyncService) SetupProxy(setup ProxySetup) error {
	// Call hotify-cli without backend-url - let hotify manage the proxy
	setupCmd := exec.Command("hotify-cli", "setup",
		"--id", setup.SiteSlug,
		"--name", setup.SiteSlug,
		"--domain", setup.Domain,
		"--port", "3100",
		"--cmd", "true", // placeholder
	)

	if output, err := setupCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to setup hotify app: %w, output: %s", err, string(output))
	}

	// Setup Traefik
	traefikCmd := exec.Command("hotify-cli", "setup-traefik",
		"--id", setup.SiteSlug,
		"--challenge-type", "http",
		"--local",
	)

	if output, err := traefikCmd.CombinedOutput(); err != nil {
		// Traefik setup is optional, log warning but don't fail
		fmt.Printf("Warning: Traefik setup failed (may require sudo): %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
	}

	return nil
}