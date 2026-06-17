package cli

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Manage users",
	Long:  `Create, read, update, and delete users for authentication and authorization.`,
}

var userEmail string
var userPassword string
var userRole string

// user list
var userListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all users",
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteUserList(target)
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

		userRepo := db.NewUserRepository()
		users, err := userRepo.List()
		if err != nil {
			fail(ExitInternal, err.Error())
		}
		writeJSON(map[string]interface{}{"version": "1.0", "users": users})
	},
}

// user create
var userCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new user",
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteUserCreate(target, cmd)
			return
		}

		if userEmail == "" {
			fail(ExitMissingFlag, "--email is required")
		}
		if userPassword == "" {
			fail(ExitMissingFlag, "--password is required")
		}
		if userRole == "" {
			userRole = "viewer"
		}

		cfg, err := config.Load()
		if err != nil {
			fail(ExitExtFailed, err.Error())
		}
		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		userRepo := db.NewUserRepository()
		user := &db.User{ID: generateID(), Email: userEmail, Role: userRole}
		if err := userRepo.Create(user, userPassword); err != nil {
			fail(ExitConflict, err.Error())
		}
		writeJSON(map[string]interface{}{
			"version": "1.0", "success": true,
			"message": "User created successfully", "user": user,
		})
	},
}

// user password
var userPasswordCmd = &cobra.Command{
	Use:   "password <email>",
	Short: "Set a user's password",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteUserPassword(target, args, cmd)
			return
		}

		if userPassword == "" {
			fail(ExitMissingFlag, "--password is required")
		}

		cfg, err := config.Load()
		if err != nil {
			fail(ExitExtFailed, err.Error())
		}
		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		userRepo := db.NewUserRepository()
		if err := userRepo.UpdatePassword(args[0], userPassword); err != nil {
			fail(ExitNotFound, err.Error())
		}
		writeJSON(map[string]interface{}{
			"version": "1.0", "success": true, "message": "Password updated successfully",
		})
	},
}

// user reset-password
var userResetPasswordCmd = &cobra.Command{
	Use:   "reset-password <email>",
	Short: "Reset a user's password to a random value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fail(ExitExtFailed, err.Error())
		}
		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		newPassword := generateRandomPassword(16)
		userRepo := db.NewUserRepository()
		if err := userRepo.UpdatePassword(args[0], newPassword); err != nil {
			fail(ExitNotFound, err.Error())
		}

		writeJSON(map[string]interface{}{
			"version": "1.0", "success": true, "message": "Password reset successfully",
			"password": newPassword,
		})
	},
}

// user grant
var userGrantCmd = &cobra.Command{
	Use:   "grant <site> <email>",
	Short: "Grant a user access to a site",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteUserGrant(target, args, cmd)
			return
		}

		siteSlug, email := args[0], args[1]
		if userRole == "" {
			userRole = "admin"
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
		site, err := siteRepo.GetBySlug(siteSlug)
		if err != nil {
			fail(ExitNotFound, "site not found")
		}

		userRepo := db.NewUserRepository()
		user, err := userRepo.GetByEmail(email)
		if err != nil {
			fail(ExitNotFound, "user not found")
		}

		if err := userRepo.GrantSiteAccess(site.ID, user.ID, userRole); err != nil {
			fail(ExitConflict, err.Error())
		}
		writeJSON(map[string]interface{}{
			"version": "1.0", "success": true,
			"message": fmt.Sprintf("Granted %s access to %s", userRole, siteSlug),
		})
	},
}

var userListRemote func(string)
var userCreateRemote func(string, *cobra.Command)
var userPasswordRemote func(string, []string, *cobra.Command)
var userGrantRemote func(string, []string, *cobra.Command)


// user revoke
var userRevokeCmd = &cobra.Command{
	Use:   "revoke <site> <email>",
	Short: "Revoke a user's access to a site",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		siteSlug, email := args[0], args[1]

		cfg, err := config.Load()
		if err != nil {
			fail(ExitExtFailed, err.Error())
		}
		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		siteRepo := db.NewSiteRepository()
		site, err := siteRepo.GetBySlug(siteSlug)
		if err != nil {
			fail(ExitNotFound, "site not found")
		}

		userRepo := db.NewUserRepository()
		user, err := userRepo.GetByEmail(email)
		if err != nil {
			fail(ExitNotFound, "user not found")
		}

		if err := userRepo.RevokeSiteAccess(site.ID, user.ID); err != nil {
			fail(ExitConflict, err.Error())
		}
		writeJSON(map[string]interface{}{
			"version": "1.0", "success": true,
			"message": fmt.Sprintf("Revoked %s access to %s", email, siteSlug),
		})
	},
}



var (
	grantBulkSites string
	grantBulkUsers string
	grantBulkRole  string
)

// user grant-bulk grants multiple users access to multiple sites
var userGrantBulkCmd = &cobra.Command{
	Use:   "grant-bulk",
	Short: "Grant multiple users access to multiple sites",
	Long:  `Grants all listed users access to all listed sites in a single command.`,
	Run: func(cmd *cobra.Command, args []string) {
		if grantBulkSites == "" || grantBulkUsers == "" {
			fail(ExitMissingFlag, "--sites and --users are required")
		}
		if grantBulkRole == "" {
			grantBulkRole = "admin"
		}

		sites := strings.Split(grantBulkSites, ",")
		users := strings.Split(grantBulkUsers, ",")

		cfg, err := config.Load()
		if err != nil {
			fail(ExitExtFailed, err.Error())
		}
		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		siteRepo := db.NewSiteRepository()
		userRepo := db.NewUserRepository()

		granted := 0
		errors := []string{}
		for _, siteSlug := range sites {
			siteSlug = strings.TrimSpace(siteSlug)
			if siteSlug == "" {
				continue
			}
			site, err := siteRepo.GetBySlug(siteSlug)
			if err != nil {
				errors = append(errors, fmt.Sprintf("site %s: not found", siteSlug))
				continue
			}
			for _, email := range users {
				email = strings.TrimSpace(email)
				if email == "" {
					continue
				}
				user, err := userRepo.GetByEmail(email)
				if err != nil {
					errors = append(errors, fmt.Sprintf("user %s: not found", email))
					continue
				}
				if err := userRepo.GrantSiteAccess(site.ID, user.ID, grantBulkRole); err != nil {
					errors = append(errors, fmt.Sprintf("%s → %s: %s", email, siteSlug, err.Error()))
				} else {
					granted++
				}
			}
		}

		result := map[string]interface{}{
			"version": "1.0", "success": len(errors) == 0,
			"granted": granted, "sites": len(sites), "users": len(users),
			"message": fmt.Sprintf("Granted %d access entries", granted),
		}
		if len(errors) > 0 {
			result["errors"] = errors
			result["message"] = fmt.Sprintf("Granted %d access entries (%d errors)", granted, len(errors))
		}
		writeJSON(result)
	},
}


func init() {
	userListCmd.Flags().String("target", "", "Remote target (host:port)")
	userCreateCmd.Flags().StringVar(&userEmail, "email", "", "User email")
	userCreateCmd.Flags().StringVar(&userPassword, "password", "", "User password")
	userCreateCmd.Flags().StringVar(&userRole, "role", "", "User role (admin, editor, viewer)")
	userCreateCmd.Flags().String("target", "", "Remote target (host:port)")
	userPasswordCmd.Flags().StringVar(&userPassword, "password", "", "New password")
	userPasswordCmd.Flags().String("target", "", "Remote target (host:port)")
	userGrantCmd.Flags().StringVar(&userRole, "role", "admin", "Role to grant (admin, editor, viewer)")
	userGrantCmd.Flags().String("target", "", "Remote target (host:port)")

	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userPasswordCmd)
	userCmd.AddCommand(userResetPasswordCmd)
	userCmd.AddCommand(userGrantCmd)
	userCmd.AddCommand(userRevokeCmd)
	userGrantBulkCmd.Flags().StringVar(&grantBulkSites, "sites", "", "Comma-separated site slugs")
	userGrantBulkCmd.Flags().StringVar(&grantBulkUsers, "users", "", "Comma-separated emails")
	userGrantBulkCmd.Flags().StringVar(&grantBulkRole, "role", "admin", "Role to grant (admin, editor, viewer)")
	userCmd.AddCommand(userGrantBulkCmd)
}

// remote handlers
func handleRemoteUserList(target string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.ListUsers()
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	writeJSON(map[string]interface{}{"version": "1.0", "users": result})
}

func handleRemoteUserCreate(target string, cmd *cobra.Command) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.CreateUser(userEmail, userPassword, userRole)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	writeJSON(map[string]interface{}{"version": "1.0", "result": result})
}

func handleRemoteUserPassword(target string, args []string, cmd *cobra.Command) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.SetUserPassword(args[0], userPassword)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	writeJSON(map[string]interface{}{"version": "1.0", "result": result})
}

func handleRemoteUserGrant(target string, args []string, cmd *cobra.Command) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fail(ExitInvalidInput, err.Error())
	}
	result, err := client.GrantSiteAccess(args[0], args[1], userRole)
	if err != nil {
		fail(ExitExtFailed, err.Error())
	}
	writeJSON(map[string]interface{}{"version": "1.0", "result": result})
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateRandomPassword(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}
