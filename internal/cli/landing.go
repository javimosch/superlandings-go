package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/javimosch/superlandings-go/internal/services"
	"github.com/spf13/cobra"
)

var landingCmd = &cobra.Command{
	Use:   "landing",
	Short: "Manage landing pages",
	Long:  `Create, read, update, and delete landing pages. Supports HTML, EJS, virtual, and static landing types.`,
}

// landing list
var landingListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all landings",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Create service
		service := services.NewLandingService(cfg)

		// List landings
		landings, err := service.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing landings: %v\n", err)
			os.Exit(1)
		}

		// Output
		if len(landings) == 0 {
			fmt.Println("No landings found")
			return
		}

		// Check if JSON output requested
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			data, err := json.Marshal(landings)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(data))
			return
		}

		// Table output
		fmt.Println("ID\tName\tSlug\tType")
		for _, landing := range landings {
			fmt.Printf("%s\t%s\t%s\t%s\n", landing.ID, landing.Name, landing.Slug, landing.Type)
		}
	},
}

// landing get
var landingGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a landing by ID or slug",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Create service
		service := services.NewLandingService(cfg)

		// Get landing
		var landing *db.Landing
		var err error

		// Try by ID first, then by slug
		landing, err = service.GetByID(args[0])
		if err != nil {
			landing, err = service.GetBySlug(args[0])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: landing not found\n")
				os.Exit(1)
			}
		}

		// Output
		jsonOutput, _ := cmd.Flags().GetBool("json")
		if jsonOutput {
			data, err := json.Marshal(landing)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(data))
			return
		}

		// Pretty output
		fmt.Printf("ID: %s\n", landing.ID)
		fmt.Printf("Name: %s\n", landing.Name)
		fmt.Printf("Slug: %s\n", landing.Slug)
		fmt.Printf("Type: %s\n", landing.Type)
		fmt.Printf("Created: %s\n", landing.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated: %s\n", landing.UpdatedAt.Format("2006-01-02 15:04:05"))
	},
}

// landing create
var landingCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new landing",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		landingType, _ := cmd.Flags().GetString("type")
		org, _ := cmd.Flags().GetString("org")
		content, _ := cmd.Flags().GetString("content")

		// Validate required fields
		if name == "" {
			fmt.Fprintf(os.Stderr, "Error: --name is required\n")
			os.Exit(1)
		}
		if slug == "" {
			fmt.Fprintf(os.Stderr, "Error: --slug is required\n")
			os.Exit(1)
		}
		if landingType == "" {
			fmt.Fprintf(os.Stderr, "Error: --type is required\n")
			os.Exit(1)
		}

		// Create service
		service := services.NewLandingService(cfg)

		// Create landing
		req := services.CreateLandingRequest{
			Name:           name,
			Slug:           slug,
			Type:           landingType,
			OrganizationID: org,
			Content:        content,
		}

		landing, err := service.Create(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating landing: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Landing created successfully!\n")
		fmt.Printf("ID: %s\n", landing.ID)
		fmt.Printf("Slug: %s\n", landing.Slug)
	},
}

// landing update
var landingUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a landing",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Get flags
		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		landingType, _ := cmd.Flags().GetString("type")
		content, _ := cmd.Flags().GetString("content")

		// Create service
		service := services.NewLandingService(cfg)

		// Update landing
		req := services.UpdateLandingRequest{
			Name:    name,
			Slug:    slug,
			Type:    landingType,
			Content: content,
		}

		landing, err := service.Update(args[0], req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating landing: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Landing updated successfully!\n")
		fmt.Printf("ID: %s\n", landing.ID)
		fmt.Printf("Slug: %s\n", landing.Slug)
	},
}

// landing delete
var landingDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a landing",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Create service
		service := services.NewLandingService(cfg)

		// Delete landing
		if err := service.Delete(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting landing: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Landing deleted successfully!\n")
	},
}

func init() {
	// Add flags
	landingListCmd.Flags().Bool("json", false, "Output as JSON")
	landingGetCmd.Flags().Bool("json", false, "Output as JSON")
	landingCreateCmd.Flags().String("name", "", "Landing name")
	landingCreateCmd.Flags().String("slug", "", "URL slug")
	landingCreateCmd.Flags().String("type", "", "Landing type (html|ejs|virtual|static)")
	landingCreateCmd.Flags().String("org", "", "Organization ID")
	landingCreateCmd.Flags().String("content", "", "HTML content (for html/ejs types)")
	landingUpdateCmd.Flags().String("name", "", "Landing name")
	landingUpdateCmd.Flags().String("slug", "", "URL slug")
	landingUpdateCmd.Flags().String("type", "", "Landing type")
	landingUpdateCmd.Flags().String("content", "", "HTML content")

	landingCmd.AddCommand(landingListCmd)
	landingCmd.AddCommand(landingGetCmd)
	landingCmd.AddCommand(landingCreateCmd)
	landingCmd.AddCommand(landingUpdateCmd)
	landingCmd.AddCommand(landingDeleteCmd)
}

// initializeDB initializes the database connection
func initializeDB() error {
	// Force usage of config package
	_ = config.Config{}
	
	if err := cfg.EnsureDirectories(); err != nil {
		return err
	}
	return db.Initialize(cfg.DatabasePath)
}