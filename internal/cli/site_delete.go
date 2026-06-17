package cli

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"path/filepath"

	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/spf13/cobra"
)

var deleteConfirm string

func init() {
	siteDeleteCmd.Flags().StringVar(&deleteConfirm, "confirm", "", "Confirmation token to execute deletion")
	siteCmd.AddCommand(siteDeleteCmd)
}

var siteDeleteCmd = &cobra.Command{
	Use:   "delete <slug>",
	Short: "Delete a site (requires --confirm token)",
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

		// Step 1: generate confirmation token
		if deleteConfirm == "" {
			token := randomToken(3)
			if err := storeDeleteToken(slug, token); err != nil {
				fail(ExitInternal, err.Error())
			}
			writeJSON(map[string]interface{}{
				"version": "1.0",
				"success": true,
				"message": fmt.Sprintf("Confirm deletion of '%s' with --confirm %s", site.Name, token),
				"token":   token,
				"hint":    fmt.Sprintf("sl-cli site delete %s --confirm %s", slug, token),
			})
			return
		}

		// Step 2: verify token and delete
		stored, err := getDeleteToken(slug)
		if err != nil || stored != deleteConfirm {
			fail(ExitInvalidInput, "invalid confirmation token — run without --confirm first to get one")
		}

		if err := repo.DeleteSite(slug); err != nil {
			fail(ExitInternal, err.Error())
		}

		// Remove site files
		os.RemoveAll(filepath.Join(cfg.SitesDir, slug))

		// Consume token
		consumeDeleteToken(slug)

		writeJSON(map[string]interface{}{
			"version": "1.0",
			"success": true,
			"message": fmt.Sprintf("Site '%s' deleted", site.Name),
		})
	},
}

func randomToken(n int) string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		b[i] = chars[idx.Int64()]
	}
	return string(b)
}

func storeDeleteToken(slug, token string) error {
	_, err := db.DB.Exec(
		`INSERT OR REPLACE INTO delete_tokens (site_slug, token, created_at) VALUES (?, ?, datetime('now'))`,
		slug, token,
	)
	return err
}

func getDeleteToken(slug string) (string, error) {
	var token string
	err := db.DB.QueryRow(`SELECT token FROM delete_tokens WHERE site_slug = ?`, slug).Scan(&token)
	return token, err
}

func consumeDeleteToken(slug string) {
	db.DB.Exec(`DELETE FROM delete_tokens WHERE site_slug = ?`, slug)
}
