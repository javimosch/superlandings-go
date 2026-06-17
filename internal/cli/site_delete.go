package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/spf13/cobra"
)

var siteDeleteCmd = &cobra.Command{
	Use:   "delete <slug>",
	Short: "Delete a site and all its files",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		slug := args[0]

		cfg, err := config.Load()
		if err != nil {
			fail(ExitInternal, err.Error())
		}

		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fail(ExitInternal, err.Error())
		}
		defer db.Close()

		repo := db.NewSiteRepository()
		site, err := repo.GetBySlug(slug)
		if err != nil {
			fail(ExitNotFound, fmt.Sprintf("site not found: %s", slug))
		}

		if err := repo.DeleteSite(slug); err != nil {
			fail(ExitInternal, err.Error())
		}

		// Remove site files from disk
		siteDir := filepath.Join(cfg.SitesDir, slug)
		os.RemoveAll(siteDir)

		writeJSON(map[string]interface{}{
			"version": "1.0",
			"success": true,
			"message": fmt.Sprintf("Site '%s' deleted", site.Name),
		})
	},
}

func init() {
	siteCmd.AddCommand(siteDeleteCmd)
}
