package cli

import (
	"fmt"
	"os"

	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/javimosch/superlandings-go/internal/services"
	"github.com/spf13/cobra"
)

// site upload
var siteUploadCmd = &cobra.Command{
	Use:   "upload <site> <path> --file <local-path>",
	Short: "Upload a shared asset (logo, CSS, JS, images)",
	Long: `Upload an asset file to the shared assets directory.
Assets are shared across all versions of a site — updating an asset
instantly affects all versions.`,
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
			"site":  args[0],
			"asset": args[1],
			"size":  len(data),
		})
	},
}

func init() {
	siteUploadCmd.Flags().String("file", "", "Local file path to upload")
	siteUploadCmd.Flags().String("target", "", "Remote target (host:port)")
	siteCmd.AddCommand(siteUploadCmd)
}
