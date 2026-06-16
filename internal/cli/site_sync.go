package cli

import (
	"fmt"
	"os"

	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/javimosch/superlandings-go/internal/services"
	"github.com/spf13/cobra"
)

// site sync
var siteSyncCmd = &cobra.Command{
	Use:   "sync <site>",
	Short: "Sync site to remote target via SSH",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		host, _ := cmd.Flags().GetString("host")
		user, _ := cmd.Flags().GetString("user")
		port, _ := cmd.Flags().GetInt("port")

		if host == "" {
			fmt.Fprintf(os.Stderr, "Error: --host is required\n")
			os.Exit(1)
		}
		if user == "" {
			user = "root"
		}

		syncService := services.NewSyncService(cfg)
		target := services.SyncTarget{
			Host: host,
			User: user,
			Port: port,
		}

		if err := syncService.Sync(args[0], target); err != nil {
			fmt.Fprintf(os.Stderr, "Error syncing site: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Site synced successfully to %s@%s\n", user, host)
	},
}

// site proxy
var siteProxyCmd = &cobra.Command{
	Use:   "proxy <site>",
	Short: "Setup hotify-cli reverse proxy for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		domain, _ := cmd.Flags().GetString("domain")
		internalURL, _ := cmd.Flags().GetString("internal-url")

		if domain == "" {
			fmt.Fprintf(os.Stderr, "Error: --domain is required\n")
			os.Exit(1)
		}
		if internalURL == "" {
			internalURL = "http://127.0.0.1:3099"
		}

		syncService := services.NewSyncService(cfg)
		setup := services.ProxySetup{
			SiteSlug:    args[0],
			Domain:      domain,
			InternalURL: internalURL,
		}

		if err := syncService.SetupProxy(setup); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting up proxy: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Proxy configured: %s -> %s\n", domain, internalURL)
	},
}

func init() {
	siteSyncCmd.Flags().String("host", "", "Remote host (e.g., 92.113.145.16)")
	siteSyncCmd.Flags().String("user", "root", "SSH user")
	siteSyncCmd.Flags().Int("port", 22, "SSH port")

	siteProxyCmd.Flags().String("domain", "", "Domain name (e.g., slv2.intrane.fr)")
	siteProxyCmd.Flags().String("internal-url", "http://127.0.0.1:3099", "Internal URL to proxy to")

	siteCmd.AddCommand(siteSyncCmd)
	siteCmd.AddCommand(siteProxyCmd)
}