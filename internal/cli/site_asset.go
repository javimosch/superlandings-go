package cli

import (
	"fmt"
	"os"

	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/javimosch/superlandings-go/internal/services"
	"github.com/spf13/cobra"
)

// site upload (create/replace)
var siteUploadCmd = &cobra.Command{
	Use:   "upload <site> <path> --file <local-path>",
	Short: "Upload a shared asset (logo, CSS, JS, images)",
	Long: `Upload an asset file to the shared assets directory.
Assets are shared across all versions of a site. Re-uploading
the same path replaces the existing asset.`,
	Example: `  sl-cli site upload my-site "logo.png" --file ./logo.png
  sl-cli site upload my-site "css/style.css" --file ./dist/style.css
  sl-cli site upload my-site "js/app.js" --file ./app.js`,
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")

		filePath, _ := cmd.Flags().GetString("file")
		if filePath == "" {
			fail(ExitMissingFlag, "--file is required")
		}

		data, err := os.ReadFile(filePath)
		if err != nil {
			fail(ExitInvalidInput, fmt.Sprintf("reading file: %v", err))
		}

		if target != "" {
			handleRemoteSiteUpload(target, args[0], args[1], data)
			return
		}

		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		service := services.NewSiteService(cfg)
		if err := service.UploadAsset(args[0], args[1], data); err != nil {
			fail(ExitNotFound, err.Error())
		}

		success("Asset uploaded", map[string]interface{}{
			"site": args[0], "asset": args[1], "size": len(data),
		})
	},
}

// site assets subcommand group
var siteAssetsCmd = &cobra.Command{
	Use:   "assets",
	Short: "Manage shared assets",
	Long:  `List and remove assets in the shared assets directory.`,
}

// site assets list
var siteAssetsListCmd = &cobra.Command{
	Use:   "list <site>",
	Short: "List all shared assets",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteAssetsList(target, args[0])
			return
		}

		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		service := services.NewSiteService(cfg)
		assets, err := service.ListAssets(args[0])
		if err != nil {
			fail(ExitNotFound, err.Error())
		}
		if assets == nil {
			assets = []services.AssetInfo{}
		}
		writeJSON(map[string]interface{}{"version": "1.0", "assets": assets})
	},
}

// site assets remove
var siteAssetsRemoveCmd = &cobra.Command{
	Use:   "remove <site> <path>",
	Short: "Remove a shared asset",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteAssetsRemove(target, args)
			return
		}

		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		service := services.NewSiteService(cfg)
		if err := service.RemoveAsset(args[0], args[1]); err != nil {
			fail(ExitNotFound, err.Error())
		}

		success("Asset removed", map[string]interface{}{
			"site":  args[0],
			"asset": args[1],
		})
	},
}

func init() {
	siteUploadCmd.Flags().String("file", "", "Local file path to upload")
	siteUploadCmd.Flags().String("target", "", "Remote target (host:port)")

	siteAssetsListCmd.Flags().String("target", "", "Remote target (host:port)")
	siteAssetsRemoveCmd.Flags().String("target", "", "Remote target (host:port)")

	siteAssetsCmd.AddCommand(siteAssetsListCmd)
	siteAssetsCmd.AddCommand(siteAssetsRemoveCmd)
	siteCmd.AddCommand(siteUploadCmd)
	siteCmd.AddCommand(siteAssetsCmd)
}
