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
			handleRemoteSiteSync(target, args[0])
			return
		}

		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		host, _ := cmd.Flags().GetString("host")
		user, _ := cmd.Flags().GetString("user")
		port, _ := cmd.Flags().GetInt("port")
		key, _ := cmd.Flags().GetString("key")

		if host == "" {
			fail(ExitMissingFlag, "--host is required when not using --target")
		}
		if user == "" {
			user = "root"
		}

		syncService := services.NewSyncService(cfg)
		syncTarget := services.SyncTarget{Host: host, User: user, Port: port, Key: key}
		if err := syncService.Sync(args[0], syncTarget); err != nil {
			fail(ExitExtFailed, err.Error())
		}

		success(fmt.Sprintf("Site synced successfully to %s@%s", user, host), nil)
	},
}

// site proxy
var siteProxyCmd = &cobra.Command{
	Use:   "proxy <site>",
	Short: "Setup hotify-cli reverse proxy for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		domain, _ := cmd.Flags().GetString("domain")
		internalURL, _ := cmd.Flags().GetString("internal-url")

		if domain == "" {
			fail(ExitMissingFlag, "--domain is required")
		}
		if internalURL == "" {
			internalURL = "http://127.0.0.1:3099"
		}

		syncService := services.NewSyncService(cfg)
		setup := services.ProxySetup{SiteSlug: args[0], Domain: domain, InternalURL: internalURL}
		if err := syncService.SetupProxy(setup); err != nil {
			fail(ExitExtFailed, err.Error())
		}

		success(fmt.Sprintf("Proxy configured: %s -> %s", domain, internalURL), map[string]interface{}{
			"domain": domain, "internal_url": internalURL,
		})
	},
}

// site import
var siteImportCmd = &cobra.Command{
	Use:   "import --input <file>",
	Short: "Import site metadata from JSON",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		input, _ := cmd.Flags().GetString("input")
		if input == "" {
			fail(ExitMissingFlag, "--input is required")
		}

		content, err := os.ReadFile(input)
		if err != nil {
			fail(ExitInvalidInput, "reading file: "+err.Error())
		}

		syncService := services.NewSyncService(cfg)
		if err := syncService.Import(string(content)); err != nil {
			fail(ExitInternal, err.Error())
		}

		success("Site imported successfully", nil)
	},
}

// site export
var siteExportCmd = &cobra.Command{
	Use:   "export <site> --output <file>",
	Short: "Export site metadata to JSON",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			output = "/tmp/site-export.json"
		}

		syncService := services.NewSyncService(cfg)
		jsonData, err := syncService.Export(args[0])
		if err != nil {
			fail(ExitNotFound, err.Error())
		}

		if err := os.WriteFile(output, []byte(jsonData), 0644); err != nil {
			fail(ExitInternal, "writing file: "+err.Error())
		}

		success(fmt.Sprintf("Site exported to %s", output), map[string]interface{}{
			"output": output,
		})
	},
}

func init() {
	siteSyncCmd.Flags().String("host", "", "Remote host")
	siteSyncCmd.Flags().String("user", "root", "SSH user")
	siteSyncCmd.Flags().Int("port", 22, "SSH port")
	siteSyncCmd.Flags().String("key", "", "SSH key path")
	siteSyncCmd.Flags().String("target", "", "Remote target name (for HTTP API sync)")

	siteProxyCmd.Flags().String("domain", "", "Domain name")
	siteProxyCmd.Flags().String("internal-url", "http://127.0.0.1:3099", "Internal URL to proxy to")

	siteImportCmd.Flags().String("input", "", "Import file path")
	siteExportCmd.Flags().String("output", "", "Output file path (default: /tmp/site-export.json)")

	siteCmd.AddCommand(siteSyncCmd)
	siteCmd.AddCommand(siteProxyCmd)
	siteCmd.AddCommand(siteImportCmd)
	siteCmd.AddCommand(siteExportCmd)
}
