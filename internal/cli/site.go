package cli

import (
	"fmt"
	"os"

	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/javimosch/superlandings-go/internal/services"
	"github.com/spf13/cobra"
)

var siteCmd = &cobra.Command{
	Use:   "site",
	Short: "Manage static sites with versioning",
	Long:  `Create and manage static sites with versioning and dynamic blocks support.`,
}

// site list
var siteListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sites",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		service := services.NewSiteService(cfg)
		sites, err := service.List()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing sites: %v\n", err)
			os.Exit(1)
		}

		if len(sites) == 0 {
			fmt.Println("No sites found")
			return
		}

		fmt.Println("ID\tName\tSlug")
		for _, site := range sites {
			fmt.Printf("%s\t%s\t%s\n", site.ID, site.Name, site.Slug)
		}
	},
}

// site create
var siteCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new site",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")

		if name == "" {
			fmt.Fprintf(os.Stderr, "Error: --name is required\n")
			os.Exit(1)
		}
		if slug == "" {
			fmt.Fprintf(os.Stderr, "Error: --slug is required\n")
			os.Exit(1)
		}

		service := services.NewSiteService(cfg)
		req := services.CreateSiteRequest{
			Name: name,
			Slug: slug,
		}

		site, err := service.Create(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating site: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Site created successfully!\n")
		fmt.Printf("ID: %s\n", site.ID)
		fmt.Printf("Slug: %s\n", site.Slug)
	},
}

// site version
var siteVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage site versions",
}

// site version create
var siteVersionCreateCmd = &cobra.Command{
	Use:   "create <site>",
	Short: "Create a new version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		version, _ := cmd.Flags().GetString("version")
		comment, _ := cmd.Flags().GetString("comment")
		author, _ := cmd.Flags().GetString("author")

		if version == "" {
			fmt.Fprintf(os.Stderr, "Error: --version is required\n")
			os.Exit(1)
		}

		service := services.NewSiteService(cfg)
		req := services.CreateVersionRequest{
			Version: version,
			Comment: comment,
			Author:  author,
		}

		createdVersion, err := service.CreateVersion(args[0], req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Version created successfully!\n")
		fmt.Printf("Version: %s\n", createdVersion.Version)
		fmt.Printf("Path: %s\n", createdVersion.Path)
		if createdVersion.IsActive {
			fmt.Println("This version is now active")
		}
	},
}

// site version list
var siteVersionListCmd = &cobra.Command{
	Use:   "list <site>",
	Short: "List all versions for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		service := services.NewSiteService(cfg)
		versions, err := service.ListVersions(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing versions: %v\n", err)
			os.Exit(1)
		}

		if len(versions) == 0 {
			fmt.Println("No versions found")
			return
		}

		fmt.Println("Version\tPath\tActive\tComment")
		for _, v := range versions {
			active := "No"
			if v.IsActive {
				active = "Yes"
			}
			fmt.Printf("%s\t%s\t%s\t%s\n", v.Version, v.Path, active, v.Comment)
		}
	},
}

// site version switch
var siteVersionSwitchCmd = &cobra.Command{
	Use:   "switch <site> <version>",
	Short: "Switch active version",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		service := services.NewSiteService(cfg)
		if err := service.SwitchVersion(args[0], args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error switching version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Switched to version %s\n", args[1])
	},
}

// site write
var siteWriteCmd = &cobra.Command{
	Use:   "write <site> <version> <file>",
	Short: "Write a file to a version",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		content, _ := cmd.Flags().GetString("content")
		if content == "" {
			fmt.Fprintf(os.Stderr, "Error: --content is required\n")
			os.Exit(1)
		}

		service := services.NewSiteService(cfg)
		if err := service.WriteFile(args[0], args[1], args[2], content); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("File written successfully: %s\n", args[2])
	},
}

func init() {
	// Site commands
	siteCreateCmd.Flags().String("name", "", "Site name")
	siteCreateCmd.Flags().String("slug", "", "Site slug")

	siteCmd.AddCommand(siteListCmd)
	siteCmd.AddCommand(siteCreateCmd)

	// Version commands
	siteVersionCreateCmd.Flags().String("version", "", "Version (e.g., v1, v2)")
	siteVersionCreateCmd.Flags().String("comment", "", "Version comment")
	siteVersionCreateCmd.Flags().String("author", "", "Author name")

	siteWriteCmd.Flags().String("content", "", "File content")

	siteVersionCmd.AddCommand(siteVersionCreateCmd)
	siteVersionCmd.AddCommand(siteVersionListCmd)
	siteVersionCmd.AddCommand(siteVersionSwitchCmd)

	siteCmd.AddCommand(siteVersionCmd)
	siteCmd.AddCommand(siteWriteCmd)
}