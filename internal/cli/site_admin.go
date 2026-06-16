package cli

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/spf13/cobra"
)

// site admin
var siteAdminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Manage site admin access",
}

// site admin create
var siteAdminCreateCmd = &cobra.Command{
	Use:   "create <site>",
	Short: "Create admin URL for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		siteSlug := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Get site by slug
		siteRepo := db.NewSiteRepository()
		site, err := siteRepo.GetBySlug(siteSlug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: site not found\n")
			os.Exit(1)
		}

		// Generate token
		token := generateRandomToken(32)

		// Set expiration (30 days from now)
		expiresAt := time.Now().Add(30 * 24 * time.Hour)

		// Create admin token
		adminRepo := db.NewSiteAdminRepository()
		if err := adminRepo.CreateAdminToken(site.ID, token, &expiresAt); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating admin token: %v\n", err)
			os.Exit(1)
		}

		adminURL := fmt.Sprintf("/admin/%s/%s", siteSlug, token)

		output := map[string]interface{}{
			"success": true,
			"admin_url": adminURL,
			"token": token,
			"expires_at": expiresAt,
		}
		jsonData, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonData))
	},
}

// site admin view
var siteAdminViewCmd = &cobra.Command{
	Use:   "view <site>",
	Short: "View admin URL for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		siteSlug := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Get site by slug
		siteRepo := db.NewSiteRepository()
		site, err := siteRepo.GetBySlug(siteSlug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: site not found\n")
			os.Exit(1)
		}

		// Get active token
		adminRepo := db.NewSiteAdminRepository()
		token, err := adminRepo.GetActiveTokenBySite(site.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "No active admin token found. Use 'sl-cli site admin create' to create one.\n")
			os.Exit(1)
		}

		adminURL := fmt.Sprintf("/admin/%s/%s", siteSlug, token.Token)

		output := map[string]interface{}{
			"success": true,
			"admin_url": adminURL,
			"token": token.Token,
			"created_at": token.CreatedAt,
			"expires_at": token.ExpiresAt,
		}
		jsonData, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonData))
	},
}

// site admin rotate
var siteAdminRotateCmd = &cobra.Command{
	Use:   "rotate <site>",
	Short: "Rotate admin token for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		siteSlug := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Get site by slug
		siteRepo := db.NewSiteRepository()
		site, err := siteRepo.GetBySlug(siteSlug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: site not found\n")
			os.Exit(1)
		}

		// Generate new token
		newToken := generateRandomToken(32)

		// Set expiration (30 days from now)
		expiresAt := time.Now().Add(30 * 24 * time.Hour)

		// Rotate token
		adminRepo := db.NewSiteAdminRepository()
		if err := adminRepo.RotateToken(site.ID, newToken, &expiresAt); err != nil {
			fmt.Fprintf(os.Stderr, "Error rotating admin token: %v\n", err)
			os.Exit(1)
		}

		adminURL := fmt.Sprintf("/admin/%s/%s", siteSlug, newToken)

		output := map[string]interface{}{
			"success": true,
			"admin_url": adminURL,
			"token": newToken,
			"expires_at": expiresAt,
		}
		jsonData, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonData))
	},
}

// site admin revoke
var siteAdminRevokeCmd = &cobra.Command{
	Use:   "revoke <site>",
	Short: "Revoke all admin tokens for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		siteSlug := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Get site by slug
		siteRepo := db.NewSiteRepository()
		site, err := siteRepo.GetBySlug(siteSlug)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: site not found\n")
			os.Exit(1)
		}

		// Revoke all tokens
		adminRepo := db.NewSiteAdminRepository()
		if err := adminRepo.RevokeAllTokens(site.ID); err != nil {
			fmt.Fprintf(os.Stderr, "Error revoking tokens: %v\n", err)
			os.Exit(1)
		}

		output := map[string]interface{}{
			"success": true,
			"message": "All admin tokens revoked",
		}
		jsonData, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonData))
	},
}

func init() {
	siteAdminCmd.AddCommand(siteAdminCreateCmd)
	siteAdminCmd.AddCommand(siteAdminViewCmd)
	siteAdminCmd.AddCommand(siteAdminRotateCmd)
	siteAdminCmd.AddCommand(siteAdminRevokeCmd)
}

func generateRandomToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}