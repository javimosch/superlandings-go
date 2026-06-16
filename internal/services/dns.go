package services

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
)

type DNSService struct {
	cfg          *config.Config
	domainRepo   *db.SiteDomainRepository
}

func NewDNSService(cfg *config.Config) *DNSService {
	return &DNSService{
		cfg:        cfg,
		domainRepo: db.NewSiteDomainRepository(),
	}
}

// SetupDNS calls hotify-cli to set up DNS for a site
func (s *DNSService) SetupDNS(siteID, siteSlug, domain, ip string, traefik bool) error {
	// Check if hotify-cli is available
	if _, err := exec.LookPath("hotify-cli"); err != nil {
		return fmt.Errorf("hotify-cli not found: %w. Install hotify-cli or configure manually", err)
	}

	// Create app in hotify-cli config with full domain (no base domain appending)
	// Use the full domain as the domain parameter to prevent duplication
	setupCmd := exec.Command("hotify-cli", "setup",
		"--id", siteSlug,
		"--name", siteSlug,
		"--domain", domain, // Use full domain
		"--port", "3099",
		"--cmd", "true", // placeholder, we only need DNS
	)

	if output, err := setupCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to setup hotify-cli app: %w, output: %s", err, string(output))
	}

	// Fix domain duplication in hotify config if needed
	// hotify-cli may append base domain, so we need to fix it
	fixCmd := exec.Command("sed", "-i",
		fmt.Sprintf("s|%s.intrane.fr|%s|g", strings.TrimSuffix(domain, ".intrane.fr"), domain),
		"~/.hotify/config.json",
	)
	fixCmd.Run() // Non-fatal if sed fails

	// Setup DNS
	dnsCmd := exec.Command("hotify-cli", "setup-dns",
		"--id", siteSlug,
		"--ip", ip,
		"--local", // use local hotify-cli
	)

	if output, err := dnsCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to setup DNS: %w, output: %s", err, string(output))
	}

	// Setup Traefik if requested
	if traefik {
		traefikCmd := exec.Command("hotify-cli", "setup-traefik",
			"--id", siteSlug,
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

// RemoveDNS removes DNS configuration via hotify-cli
func (s *DNSService) RemoveDNS(siteSlug string) error {
	// Check if hotify-cli is available
	if _, err := exec.LookPath("hotify-cli"); err != nil {
		return fmt.Errorf("hotify-cli not found: %w. Remove DNS configuration manually", err)
	}

	// Prune DNS/Traefik
	pruneCmd := exec.Command("hotify-cli", "prune",
		"--id", siteSlug,
		"--local",
	)

	if output, err := pruneCmd.CombinedOutput(); err != nil {
		// Non-fatal, might not exist
		fmt.Printf("Warning: failed to prune DNS/Traefik: %v, output: %s\n", err, string(output))
	}

	// Remove app from hotify-cli
	removeCmd := exec.Command("hotify-cli", "remove",
		"--id", siteSlug,
	)

	if output, err := removeCmd.CombinedOutput(); err != nil {
		// Non-fatal, app might not exist
		fmt.Printf("Warning: failed to remove hotify-cli app: %v, output: %s\n", err, string(output))
	}

	return nil
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
	checkCmd := exec.Command("hotify-cli", "list", "--json")
	output, err := checkCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check hotify-cli apps: %w", err)
	}

	// Check if domain is already used (simple string check)
	if strings.Contains(string(output), domain) {
		return fmt.Errorf("domain already exists in hotify-cli")
	}

	return nil
}