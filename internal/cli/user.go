package cli

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

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
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		if err := db.Initialize(cfg.DatabasePath); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		userRepo := db.NewUserRepository()
		users, err := userRepo.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing users: %v\n", err)
			os.Exit(1)
		}

		output := map[string]interface{}{"users": users}
		jsonData, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonData))
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
			fmt.Fprintf(os.Stderr, "Error: --email is required\n")
			os.Exit(1)
		}
		if userPassword == "" {
			fmt.Fprintf(os.Stderr, "Error: --password is required\n")
			os.Exit(1)
		}
		if userRole == "" {
			userRole = "viewer"
		}

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

		userRepo := db.NewUserRepository()
		user := &db.User{
			ID:    generateID(),
			Email: userEmail,
			Role:  userRole,
		}

		if err := userRepo.Create(user, userPassword); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating user: %v\n", err)
			os.Exit(1)
		}

		output := map[string]interface{}{
			"success": true,
			"user":    user,
		}
		jsonData, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonData))
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

		email := args[0]
		if userPassword == "" {
			fmt.Fprintf(os.Stderr, "Error: --password is required\n")
			os.Exit(1)
		}

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

		userRepo := db.NewUserRepository()
		if err := userRepo.UpdatePassword(email, userPassword); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating password: %v\n", err)
			os.Exit(1)
		}

		output := map[string]interface{}{
			"success": true,
			"message": "Password updated successfully",
		}
		jsonData, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonData))
	},
}

// user reset-password
var userResetPasswordCmd = &cobra.Command{
	Use:   "reset-password <email>",
	Short: "Reset a user's password to a random value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		email := args[0]

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

		// Generate random password
		newPassword := generateRandomPassword(16)

		userRepo := db.NewUserRepository()
		if err := userRepo.UpdatePassword(email, newPassword); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating password: %v\n", err)
			os.Exit(1)
		}

		output := map[string]interface{}{
			"success":  true,
			"message":  "Password reset successfully",
			"password": newPassword,
		}
		jsonData, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonData))
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

		siteSlug := args[0]
		email := args[1]
		if userRole == "" {
			userRole = "viewer"
		}

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

		// Get user by email
		userRepo := db.NewUserRepository()
		user, err := userRepo.GetByEmail(email)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: user not found\n")
			os.Exit(1)
		}

		// Grant access
		if err := userRepo.GrantSiteAccess(site.ID, user.ID, userRole); err != nil {
			fmt.Fprintf(os.Stderr, "Error granting access: %v\n", err)
			os.Exit(1)
		}

		output := map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Granted %s access to %s", userRole, siteSlug),
		}
		jsonData, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(jsonData))
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

	userGrantCmd.Flags().StringVar(&userRole, "role", "", "Role to grant (editor, viewer)")
	userGrantCmd.Flags().String("target", "", "Remote target (host:port)")

	userCmd.AddCommand(userListCmd)
	userCmd.AddCommand(userCreateCmd)
	userCmd.AddCommand(userPasswordCmd)
	userCmd.AddCommand(userResetPasswordCmd)
	userCmd.AddCommand(userGrantCmd)
}

func handleRemoteUserList(target string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	result, err := client.ListUsers()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonData))
}

func handleRemoteUserCreate(target string, cmd *cobra.Command) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	result, err := client.CreateUser(userEmail, userPassword, userRole)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonData))
}

func handleRemoteUserPassword(target string, args []string, cmd *cobra.Command) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	email := args[0]
	result, err := client.SetUserPassword(email, userPassword)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonData))
}

func handleRemoteUserGrant(target string, args []string, cmd *cobra.Command) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	siteSlug := args[0]
	email := args[1]
	result, err := client.GrantSiteAccess(siteSlug, email, userRole)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	jsonData, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonData))
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