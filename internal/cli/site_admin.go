package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteSiteAdminCreate(target, args[0])
			return
		}

		cfg, err := config.Load()
		if err != nil {
			fail(ExitExtFailed, err.Error())
		}
		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		siteRepo := db.NewSiteRepository()
		site, err := siteRepo.GetBySlug(args[0])
		if err != nil {
			fail(ExitNotFound, "site not found")
		}

		token := generateRandomToken(32)
		expiresAt := time.Now().Add(30 * 24 * time.Hour)
		adminRepo := db.NewSiteAdminRepository()
		if err := adminRepo.CreateAdminToken(site.ID, token, &expiresAt); err != nil {
			fail(ExitInternal, err.Error())
		}

		writeJSON(map[string]interface{}{
			"version": "1.0", "success": true,
			"admin_url": fmt.Sprintf("/admin/%s/%s", args[0], token),
			"token":     token, "expires_at": expiresAt,
		})
	},
}

// site admin view
var siteAdminViewCmd = &cobra.Command{
	Use:   "view <site>",
	Short: "View admin URL for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteSiteAdminView(target, args[0])
			return
		}

		cfg, err := config.Load()
		if err != nil {
			fail(ExitExtFailed, err.Error())
		}
		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		siteRepo := db.NewSiteRepository()
		site, err := siteRepo.GetBySlug(args[0])
		if err != nil {
			fail(ExitNotFound, "site not found")
		}

		adminRepo := db.NewSiteAdminRepository()
		adminToken, err := adminRepo.GetActiveTokenBySite(site.ID)
		if err != nil {
			fail(ExitNotFound, "No active admin token found. Use 'sl-cli site admin create' to create one.")
		}

		writeJSON(map[string]interface{}{
			"version": "1.0", "success": true,
			"admin_url": fmt.Sprintf("/admin/%s/%s", args[0], adminToken.Token),
			"token":     adminToken.Token,
			"created_at": adminToken.CreatedAt,
			"expires_at": adminToken.ExpiresAt,
		})
	},
}

// site admin rotate
var siteAdminRotateCmd = &cobra.Command{
	Use:   "rotate <site>",
	Short: "Rotate admin token for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteSiteAdminRotate(target, args[0])
			return
		}

		cfg, err := config.Load()
		if err != nil {
			fail(ExitExtFailed, err.Error())
		}
		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		siteRepo := db.NewSiteRepository()
		site, err := siteRepo.GetBySlug(args[0])
		if err != nil {
			fail(ExitNotFound, "site not found")
		}

		newToken := generateRandomToken(32)
		expiresAt := time.Now().Add(30 * 24 * time.Hour)
		adminRepo := db.NewSiteAdminRepository()
		if err := adminRepo.RotateToken(site.ID, newToken, &expiresAt); err != nil {
			fail(ExitInternal, err.Error())
		}

		writeJSON(map[string]interface{}{
			"version": "1.0", "success": true,
			"admin_url": fmt.Sprintf("/admin/%s/%s", args[0], newToken),
			"token":     newToken, "expires_at": expiresAt,
		})
	},
}

// site admin revoke
var siteAdminRevokeCmd = &cobra.Command{
	Use:   "revoke <site>",
	Short: "Revoke all admin tokens for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteSiteAdminRevoke(target, args[0])
			return
		}

		cfg, err := config.Load()
		if err != nil {
			fail(ExitExtFailed, err.Error())
		}
		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		siteRepo := db.NewSiteRepository()
		site, err := siteRepo.GetBySlug(args[0])
		if err != nil {
			fail(ExitNotFound, "site not found")
		}

		adminRepo := db.NewSiteAdminRepository()
		if err := adminRepo.RevokeAllTokens(site.ID); err != nil {
			fail(ExitInternal, err.Error())
		}

		writeJSON(map[string]interface{}{
			"version": "1.0", "success": true, "message": "All admin tokens revoked",
		})
	},
}

func init() {
	siteAdminCreateCmd.Flags().String("target", "", "Remote target (host:port)")
	siteAdminViewCmd.Flags().String("target", "", "Remote target (host:port)")
	siteAdminRotateCmd.Flags().String("target", "", "Remote target (host:port)")
	siteAdminRevokeCmd.Flags().String("target", "", "Remote target (host:port)")

	siteAdminCmd.AddCommand(siteAdminCreateCmd)
	siteAdminCmd.AddCommand(siteAdminViewCmd)
	siteAdminCmd.AddCommand(siteAdminRotateCmd)
	siteAdminCmd.AddCommand(siteAdminRevokeCmd)
}

func handleRemoteSiteAdminCreate(target, siteSlug string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.CreateSiteAdminToken(siteSlug)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	writeJSON(map[string]interface{}{"version": "1.0", "result": result})
}

func handleRemoteSiteAdminView(target, siteSlug string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.GetSiteAdminToken(siteSlug)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	writeJSON(map[string]interface{}{"version": "1.0", "result": result})
}

func handleRemoteSiteAdminRotate(target, siteSlug string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.RotateSiteAdminToken(siteSlug)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	writeJSON(map[string]interface{}{"version": "1.0", "result": result})
}

func handleRemoteSiteAdminRevoke(target, siteSlug string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.RevokeSiteAdminToken(siteSlug)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	writeJSON(map[string]interface{}{"version": "1.0", "result": result})
}

func generateRandomToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}
