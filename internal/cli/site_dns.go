package cli

import (
	"fmt"

	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/javimosch/superlandings-go/internal/services"
	"github.com/spf13/cobra"
)

// site dns
var siteDnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "Manage DNS via hotify-cli",
}

// site dns setup
var siteDnsSetupCmd = &cobra.Command{
	Use:   "setup <site>",
	Short: "Setup DNS for a site via hotify-cli",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteDNSSetup(target, args, cmd)
			return
		}

		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		domain, _ := cmd.Flags().GetString("domain")
		ip, _ := cmd.Flags().GetString("ip")
		traefik, _ := cmd.Flags().GetBool("traefik")

		if domain == "" {
			fail(ExitMissingFlag, "--domain is required")
		}
		if ip == "" {
			fail(ExitMissingFlag, "--ip is required")
		}

		siteService := services.NewSiteService(cfg)
		site, err := siteService.GetBySlug(args[0])
		if err != nil {
			fail(ExitNotFound, "site not found")
		}

		dnsService := services.NewDNSService(cfg)
		if err := dnsService.SetupDNS(site.ID, site.Slug, domain, ip, traefik); err != nil {
			fail(ExitExtFailed, err.Error())
		}

		success(fmt.Sprintf("DNS setup successfully for %s -> %s", domain, ip), map[string]interface{}{
			"domain": domain, "ip": ip, "traefik": traefik,
		})
	},
}

// site dns list
var siteDnsListCmd = &cobra.Command{
	Use:   "list <site>",
	Short: "List DNS domains for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteDNSList(target, args[0])
			return
		}

		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		siteService := services.NewSiteService(cfg)
		site, err := siteService.GetBySlug(args[0])
		if err != nil {
			fail(ExitNotFound, "site not found")
		}

		dnsService := services.NewDNSService(cfg)
		domains, err := dnsService.GetDomains(site.ID)
		if err != nil {
			fail(ExitInternal, err.Error())
		}
		if domains == nil {
			domains = []db.SiteDomain{}
		}

		writeJSON(map[string]interface{}{"version": "1.0", "domains": domains})
	},
}

// site dns remove
var siteDnsRemoveCmd = &cobra.Command{
	Use:   "remove <site>",
	Short: "Remove DNS configuration for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteDNSRemove(target, args)
			return
		}

		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		dnsService := services.NewDNSService(cfg)
		if err := dnsService.RemoveDNS(args[0]); err != nil {
			fail(ExitExtFailed, err.Error())
		}

		success(fmt.Sprintf("DNS configuration removed for %s", args[0]), nil)
	},
}

func init() {
	siteDnsSetupCmd.Flags().String("domain", "", "Domain name")
	siteDnsSetupCmd.Flags().String("ip", "", "IP address")
	siteDnsSetupCmd.Flags().Bool("traefik", false, "Setup Traefik routing")
	siteDnsSetupCmd.Flags().String("target", "", "Remote target (host:port)")
	siteDnsListCmd.Flags().String("target", "", "Remote target (host:port)")
	siteDnsRemoveCmd.Flags().String("target", "", "Remote target (host:port)")

	siteDnsCmd.AddCommand(siteDnsSetupCmd)
	siteDnsCmd.AddCommand(siteDnsListCmd)
	siteDnsCmd.AddCommand(siteDnsRemoveCmd)
}
