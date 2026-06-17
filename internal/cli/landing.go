package cli

import (
	"encoding/json"

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
		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		service := services.NewLandingService(cfg)
		landings, err := service.List()
		if err != nil {
			fail(ExitInternal, err.Error())
		}
		if landings == nil {
			landings = []db.Landing{}
		}
		data, _ := json.Marshal(landings)
		out := map[string]interface{}{
			"version":  "1.0",
			"landings": json.RawMessage(data),
		}
		writeJSON(out)
	},
}

// landing get
var landingGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a landing by ID or slug",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		service := services.NewLandingService(cfg)
		landing, err := service.GetByID(args[0])
		if err != nil {
			landing, err = service.GetBySlug(args[0])
			if err != nil {
				fail(ExitNotFound, "landing not found")
			}
		}

		data, _ := json.Marshal(landing)
		out := map[string]interface{}{
			"version": "1.0",
			"landing": json.RawMessage(data),
		}
		writeJSON(out)
	},
}

// landing create
var landingCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new landing",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		landingType, _ := cmd.Flags().GetString("type")
		org, _ := cmd.Flags().GetString("org")
		content, _ := cmd.Flags().GetString("content")

		if name == "" {
			fail(ExitMissingFlag, "--name is required")
		}
		if slug == "" {
			fail(ExitMissingFlag, "--slug is required")
		}
		if landingType == "" {
			fail(ExitMissingFlag, "--type is required")
		}

		service := services.NewLandingService(cfg)
		req := services.CreateLandingRequest{
			Name: name, Slug: slug, Type: landingType,
			OrganizationID: org, Content: content,
		}
		landing, err := service.Create(req)
		if err != nil {
			fail(ExitConflict, err.Error())
		}

		success("Landing created successfully", map[string]interface{}{
			"id": landing.ID, "slug": landing.Slug,
		})
	},
}

// landing update
var landingUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a landing",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		landingType, _ := cmd.Flags().GetString("type")
		content, _ := cmd.Flags().GetString("content")

		service := services.NewLandingService(cfg)
		req := services.UpdateLandingRequest{Name: name, Slug: slug, Type: landingType, Content: content}
		landing, err := service.Update(args[0], req)
		if err != nil {
			fail(ExitNotFound, err.Error())
		}

		success("Landing updated successfully", map[string]interface{}{
			"id": landing.ID, "slug": landing.Slug,
		})
	},
}

// landing delete
var landingDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a landing",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		service := services.NewLandingService(cfg)
		if err := service.Delete(args[0]); err != nil {
			fail(ExitNotFound, err.Error())
		}

		success("Landing deleted successfully", nil)
	},
}

func init() {
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
