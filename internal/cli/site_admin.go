package cli

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/javimosch/superlandings-go/internal/services"
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
		adminRepo := db.NewSiteAdminRepository()
		if err := adminRepo.CreateAdminToken(site.ID, token, nil); err != nil {
			fail(ExitInternal, err.Error())
		}

		writeJSON(map[string]interface{}{
			"version": "1.0", "success": true,
			"admin_url": fmt.Sprintf("/admin/%s/%s", args[0], token),
			"token": token,
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
		adminRepo := db.NewSiteAdminRepository()
		if err := adminRepo.RotateToken(site.ID, newToken, nil); err != nil {
			fail(ExitInternal, err.Error())
		}

		writeJSON(map[string]interface{}{
			"version": "1.0", "success": true,
			"admin_url": fmt.Sprintf("/admin/%s/%s", args[0], newToken),
			"token":     newToken,
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

// site admin configure
var siteAdminConfigureCmd = &cobra.Command{
	Use:   "configure <site> --auto-detect",
	Short: "Configure admin panel via auto-detection or schema",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		service := services.NewSiteService(cfg)
		site, err := service.GetBySlug(args[0])
		if err != nil {
			fail(ExitNotFound, "site not found")
		}

		schema := map[string]interface{}{
			"version": "1.0",
			"auth":    "none",
		}
		sections := []map[string]interface{}{}

		// Auto-detect: check for data files
		version, _ := service.GetVersionBySiteAndVersion(site.ID, "")
		versionDir := filepath.Join(cfg.SitesDir, site.Slug)
		if version != nil {
			versionDir = filepath.Join(versionDir, version.Version)
		} else {
			versionDir = filepath.Join(versionDir, "v1")
		}

		// Detect .data.json files → form sections
		if entries, err := os.ReadDir(versionDir); err == nil {
			for _, e := range entries {
				if strings.HasSuffix(e.Name(), ".data.json") {
					dataPath := filepath.Join(versionDir, e.Name())
					if data, err := os.ReadFile(dataPath); err == nil {
						var fields map[string]interface{}
						if json.Unmarshal(data, &fields) == nil {
							fieldDefs := []map[string]interface{}{}
							for k, v := range fields {
								ft := "text"
								if _, ok := v.(string); ok {
									if len(v.(string)) > 80 {
										ft = "textarea"
									}
								} else if _, ok := v.(float64); ok {
									ft = "number"
								} else if _, ok := v.(bool); ok {
									ft = "toggle"
								}
								fieldDefs = append(fieldDefs, map[string]interface{}{
									"key": k, "label": strings.Title(strings.ReplaceAll(k, "_", " ")),
									"type": ft,
								})
							}
							sections = append(sections, map[string]interface{}{
								"id":     "data",
								"type":   "form",
								"title":  "Site Data",
								"source": e.Name(),
								"fields": fieldDefs,
							})
						}
					}
					break // one form section is enough
				}
			}
		}

		// Detect blog/ directory → markdown section
		blogDir := filepath.Join(cfg.SitesDir, site.Slug, "assets")
		_ = blogDir
		blogContentDir := filepath.Join(versionDir, "blog")
		if _, err := os.Stat(blogContentDir); err == nil {
			sections = append(sections, map[string]interface{}{
				"id":     "blog",
				"type":   "markdown",
				"title":  "Blog Posts",
				"source": "blog/",
			})
		}

		schema["sections"] = sections

		// Write schema to site directory
		schemaPath := filepath.Join(cfg.SitesDir, site.Slug, "admin-schema.json")
		schemaJSON, _ := json.MarshalIndent(schema, "", "  ")
		if err := os.WriteFile(schemaPath, schemaJSON, 0644); err != nil {
			fail(ExitInternal, "writing schema: "+err.Error())
		}

		writeJSON(schema)
	},
}

func init() {
	siteAdminConfigureCmd.Flags().Bool("auto-detect", false, "Auto-detect site structure and generate schema")

	siteAdminCreateCmd.Flags().String("target", "", "Remote target (host:port)")
	siteAdminViewCmd.Flags().String("target", "", "Remote target (host:port)")
	siteAdminRotateCmd.Flags().String("target", "", "Remote target (host:port)")
	siteAdminRevokeCmd.Flags().String("target", "", "Remote target (host:port)")

	siteAdminCmd.AddCommand(siteAdminCreateCmd)
	siteAdminCmd.AddCommand(siteAdminViewCmd)
	siteAdminCmd.AddCommand(siteAdminRotateCmd)
	siteAdminCmd.AddCommand(siteAdminRevokeCmd)
	siteAdminCmd.AddCommand(siteAdminConfigureCmd)
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
