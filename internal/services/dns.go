package services

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
)

type DNSService struct {
	cfg              *config.Config
	domainRepo       *db.SiteDomainRepository
	hotifyConfigPath string // defaults to ~/.hotify/config.json
}

func NewDNSService(cfg *config.Config) *DNSService {
	return &DNSService{
		cfg:        cfg,
		domainRepo: db.NewSiteDomainRepository(),
	}
}

// hotifyAppIDPrefix namespaces the apps this service creates in hotify-cli.
// The id used to be the bare site slug, and hotify's setup is an upsert: a
// site named "mago" would have silently replaced the unrelated, live hotify
// app "mago" (found in the go.mago.intrane.fr pilot, 2026-09-24).
const hotifyAppIDPrefix = "sl-"

type hotifyApp struct {
	ID     string `json:"id"`
	Domain string `json:"domain"`
	Port   int    `json:"port"`
}

// hotifyApps reads the apps from hotify-cli's config. `hotify-cli list` has
// no JSON output, so the config file is the only machine-readable source.
func (s *DNSService) hotifyApps() ([]hotifyApp, error) {
	path := s.hotifyConfigPath
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(home, ".hotify", "config.json")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading hotify config: %w", err)
	}
	var cfg struct {
		Apps []hotifyApp `json:"apps"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing hotify config %s: %w", path, err)
	}
	return cfg.Apps, nil
}

// resolveHotifyApp picks the hotify app id for a site served on port, and
// refuses when that would touch an app this backend does not own.
// A legacy app named after the bare slug is adopted only when it already
// points at this backend's port -- otherwise it belongs to someone else.
func resolveHotifyApp(apps []hotifyApp, siteSlug, domain string, port int) (id string, exists bool, err error) {
	id = hotifyAppIDPrefix + siteSlug
	for _, a := range apps {
		if a.ID == siteSlug && a.Port == port {
			id = siteSlug
		}
	}
	for _, a := range apps {
		switch {
		case a.ID == id && a.Port != port:
			return "", false, fmt.Errorf("hotify app %q exists on port %d, not this backend's %d; refusing to overwrite it", a.ID, a.Port, port)
		case a.ID == id:
			exists = true
		case domain != "" && a.Domain == domain:
			return "", false, fmt.Errorf("domain %s is already served by hotify app %q (port %d)", domain, a.ID, a.Port)
		}
	}
	return id, exists, nil
}

// SetupDNS calls hotify-cli to set up DNS for a site served by this backend on port.
func (s *DNSService) SetupDNS(siteID, siteSlug, domain, ip string, port int, traefik bool) error {
	// Check if hotify-cli is available
	if _, err := exec.LookPath("hotify-cli"); err != nil {
		return fmt.Errorf("hotify-cli not found: %w. Install hotify-cli or configure manually", err)
	}
	if port <= 0 {
		return fmt.Errorf("backend port is required")
	}

	apps, err := s.hotifyApps()
	if err != nil {
		return err
	}
	appID, _, err := resolveHotifyApp(apps, siteSlug, domain, port)
	if err != nil {
		return err
	}

	// Create app in hotify-cli config with full domain
	setupCmd := exec.Command("hotify-cli", "setup",
		"--id", appID,
		"--name", appID,
		"--domain", domain,
		"--port", strconv.Itoa(port),
		"--cmd", "true", // placeholder, we only need DNS
	)

	if output, err := setupCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to setup hotify-cli app: %w, output: %s", err, string(output))
	}

	// hotify-cli is not supposed to rewrite the domain; check rather than
	// patch its config behind its back.
	if apps, err = s.hotifyApps(); err != nil {
		return err
	}
	for _, a := range apps {
		if a.ID == appID && a.Domain != domain {
			return fmt.Errorf("hotify app %q was saved with domain %q, expected %q", appID, a.Domain, domain)
		}
	}

	// Setup DNS
	dnsCmd := exec.Command("hotify-cli", "setup-dns",
		"--id", appID,
		"--ip", ip,
		"--local", // use local hotify-cli
	)

	if output, err := dnsCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to setup DNS: %w, output: %s", err, string(output))
	}

	// Setup Traefik if requested
	if traefik {
		traefikCmd := exec.Command("hotify-cli", "setup-traefik",
			"--id", appID,
			"--challenge-type", "http",
			"--local",
		)

		if output, err := traefikCmd.CombinedOutput(); err != nil {
			// Traefik setup is optional - don't fail the whole operation
			fmt.Printf("Warning: Failed to setup Traefik: %v, output: %s\n", err, string(output))
			fmt.Printf("DNS is configured but SSL certificates may need manual setup\n")
		}
	}

	// Save domain to database
	domainRecord := &db.SiteDomain{
		SiteID:  siteID,
		Domain:  domain,
		IP:      ip,
		Traefik: traefik,
	}

	if err := s.domainRepo.Create(domainRecord); err != nil {
		return fmt.Errorf("failed to save domain to database: %w", err)
	}

	return nil
}

// RemoveDNS removes DNS configuration via hotify-cli, for the hotify app this
// backend (on port) created for the site. It never prunes an app pointing at
// another port: with the bare slug as id, removing site "mago" would have
// taken down the live mago app.
func (s *DNSService) RemoveDNS(siteSlug string, port int) (removedApp string, err error) {
	// Check if hotify-cli is available
	if _, err := exec.LookPath("hotify-cli"); err != nil {
		return "", fmt.Errorf("hotify-cli not found: %w. Remove DNS configuration manually", err)
	}

	apps, err := s.hotifyApps()
	if err != nil {
		return "", err
	}
	appID, exists, err := resolveHotifyApp(apps, siteSlug, "", port)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", nil
	}

	// Prune DNS/Traefik
	pruneCmd := exec.Command("hotify-cli", "prune",
		"--id", appID,
		"--local",
	)

	if output, err := pruneCmd.CombinedOutput(); err != nil {
		// Non-fatal, might not exist
		fmt.Printf("Warning: failed to prune DNS/Traefik: %v, output: %s\n", err, string(output))
	}

	// Remove app from hotify-cli
	removeCmd := exec.Command("hotify-cli", "remove",
		"--id", appID,
	)

	if output, err := removeCmd.CombinedOutput(); err != nil {
		// Non-fatal, app might not exist
		fmt.Printf("Warning: failed to remove hotify-cli app: %v, output: %s\n", err, string(output))
	}

	return appID, nil
}

// GetDomains returns all domains for a site
func (s *DNSService) GetDomains(siteID string) ([]db.SiteDomain, error) {
	return s.domainRepo.GetBySiteID(siteID)
}

// GetDomainByDomain returns a domain by its name
func (s *DNSService) GetDomainByDomain(domain string) (*db.SiteDomain, error) {
	return s.domainRepo.GetByDomain(domain)
}

// ValidateDomain checks if a domain is valid
func (s *DNSService) ValidateDomain(domain string) error {
	if domain == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Basic domain validation
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("invalid domain format")
	}

	// Check if domain already exists in hotify-cli
	apps, err := s.hotifyApps()
	if err != nil {
		return fmt.Errorf("failed to check hotify-cli apps: %w", err)
	}
	for _, a := range apps {
		if a.Domain == domain {
			return fmt.Errorf("domain already exists in hotify-cli (app %q)", a.ID)
		}
	}

	return nil
}