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
	Short: "Sync site to remote target via SSH or HTTP API",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		
		if target != "" {
			// Remote HTTP sync
			handleRemoteSiteSync(target, args[0])
			return
		}
		
		// Local SSH sync (existing implementation)
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		host, _ := cmd.Flags().GetString("host")
		user, _ := cmd.Flags().GetString("user")
		port, _ := cmd.Flags().GetInt("port")
		key, _ := cmd.Flags().GetString("key")

		if host == "" {
			fmt.Fprintf(os.Stderr, "Error: --host is required when not using --target\n")
			os.Exit(1)
		}
		if user == "" {
			user = "root"
		}

		syncService := services.NewSyncService(cfg)
		syncTarget := services.SyncTarget{
			Host: host,
			User: user,
			Port: port,
			Key:  key,
		}

		if err := syncService.Sync(args[0], syncTarget); err != nil {
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

// site import
var siteImportCmd = &cobra.Command{
	Use:   "import --input <file>",
	Short: "Import site metadata from JSON",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		input, _ := cmd.Flags().GetString("input")
		if input == "" {
			fmt.Fprintf(os.Stderr, "Error: --input is required\n")
			os.Exit(1)
		}

		content, err := os.ReadFile(input)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}

		syncService := services.NewSyncService(cfg)
		if err := syncService.Import(string(content)); err != nil {
			fmt.Fprintf(os.Stderr, "Error importing: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Site imported successfully")
	},
}

// site export
var siteExportCmd = &cobra.Command{
	Use:   "export <site> --output <file>",
	Short: "Export site metadata to JSON",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = "/tmp/site-export.json"
		}

		syncService := services.NewSyncService(cfg)
		jsonData, err := syncService.Export(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error exporting: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(output, []byte(jsonData), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Site exported to %s\n", output)
	},
}

func init() {
	siteSyncCmd.Flags().String("host", "", "Remote host (e.g., 92.113.145.16)")
	siteSyncCmd.Flags().String("user", "root", "SSH user")
	siteSyncCmd.Flags().Int("port", 22, "SSH port")
	siteSyncCmd.Flags().String("key", "", "SSH key path (e.g., ~/.ssh/id_rsa_srv)")
	siteSyncCmd.Flags().String("target", "", "Remote target name (for HTTP API sync)")

	siteProxyCmd.Flags().String("domain", "", "Domain name (e.g., slv2.intrane.fr)")
	siteProxyCmd.Flags().String("internal-url", "http://127.0.0.1:3099", "Internal URL to proxy to")

	siteImportCmd.Flags().String("input", "", "Import file path")

	siteExportCmd.Flags().String("output", "", "Output file path (default: /tmp/site-export.json)")

	siteCmd.AddCommand(siteSyncCmd)
	siteCmd.AddCommand(siteProxyCmd)
	siteCmd.AddCommand(siteImportCmd)
	siteCmd.AddCommand(siteExportCmd)
}