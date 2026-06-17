package cli

import (
	"fmt"

	"github.com/javimosch/superlandings-go/internal/config"
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
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteSiteList(target)
			return
		}

		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		service := services.NewSiteService(cfg)
		sites, err := service.List()
		if err != nil {
			fail(ExitInternal, err.Error())
		}

		type siteEntry struct {
			ID      string   `json:"id"`
			Name    string   `json:"name"`
			Slug    string   `json:"slug"`
			Domains []string `json:"domains"`
		}

		dnsService := services.NewDNSService(cfg)
		entries := make([]siteEntry, 0, len(sites))
		for _, site := range sites {
			entry := siteEntry{ID: site.ID, Name: site.Name, Slug: site.Slug}
			domains, err := dnsService.GetDomains(site.ID)
			if err == nil {
				for _, d := range domains {
					entry.Domains = append(entry.Domains, d.Domain)
				}
			}
			entries = append(entries, entry)
		}

		writeJSON(map[string]interface{}{"version": "1.0", "sites": entries})
	},
}

// site create
var siteCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new site",
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")

		if name == "" {
			fail(ExitMissingFlag, "--name is required")
		}
		if slug == "" {
			fail(ExitMissingFlag, "--slug is required")
		}

		service := services.NewSiteService(cfg)
		req := services.CreateSiteRequest{Name: name, Slug: slug}
		site, err := service.Create(req)
		if err != nil {
			fail(ExitConflict, err.Error())
		}

		success("Site created successfully", map[string]interface{}{
			"id": site.ID, "slug": site.Slug,
		})
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
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteVersionCreate(target, args, cmd)
			return
		}

		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		version, _ := cmd.Flags().GetString("version")
		comment, _ := cmd.Flags().GetString("comment")
		author, _ := cmd.Flags().GetString("author")

		if version == "" {
			fail(ExitMissingFlag, "--version is required")
		}

		service := services.NewSiteService(cfg)
		req := services.CreateVersionRequest{Version: version, Comment: comment, Author: author}
		createdVersion, err := service.CreateVersion(args[0], req)
		if err != nil {
			fail(ExitNotFound, err.Error())
		}

		success("Version created successfully", map[string]interface{}{
			"version": createdVersion.Version,
			"path":    createdVersion.Path,
			"isActive": createdVersion.IsActive,
		})
	},
}

// site version list
var siteVersionListCmd = &cobra.Command{
	Use:   "list <site>",
	Short: "List all versions for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteVersionList(target, args[0])
			return
		}

		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		service := services.NewSiteService(cfg)
		versions, err := service.ListVersions(args[0])
		if err != nil {
			fail(ExitNotFound, err.Error())
		}
		if versions == nil {
			versions = []db.SiteVersion{}
		}

		writeJSON(map[string]interface{}{"version": "1.0", "versions": versions})
	},
}

// site version switch
var siteVersionSwitchCmd = &cobra.Command{
	Use:   "switch <site> <version>",
	Short: "Switch active version",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteVersionSwitch(target, args[0], args[1])
			return
		}

		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		service := services.NewSiteService(cfg)
		if err := service.SwitchVersion(args[0], args[1]); err != nil {
			fail(ExitNotFound, err.Error())
		}

		success(fmt.Sprintf("Switched to version %s", args[1]), map[string]interface{}{
			"version": args[1], "site": args[0],
		})
	},
}

// site write
var siteWriteCmd = &cobra.Command{
	Use:   "write <site> <version> <file>",
	Short: "Write a file to a version",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		target, _ := cmd.Flags().GetString("target")
		if target != "" {
			handleRemoteSiteWrite(target, args, cmd)
			return
		}

		if err := initializeDB(); err != nil {
			fail(ExitExtFailed, "database init: "+err.Error())
		}
		defer db.Close()

		content, _ := cmd.Flags().GetString("content")
		if content == "" {
			fail(ExitMissingFlag, "--content is required")
		}

		service := services.NewSiteService(cfg)
		if err := service.WriteFile(args[0], args[1], args[2], content); err != nil {
			fail(ExitNotFound, err.Error())
		}

		success("File written successfully", map[string]interface{}{
			"file": args[2], "site": args[0], "version": args[1],
		})
	},
}

var (
	siteSnapshotName string
	siteVersionTarget string
)

// site version snapshot
var siteVersionSnapshotCmd = &cobra.Command{
	Use:   "snapshot <site>",
	Short: "Create a named snapshot of the active version",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _ := config.Load()
		db.Initialize(cfg.DatabasePath)
		defer db.Close()

		svc := services.NewVersioningService(cfg)
		ver, err := svc.Snapshot(args[0], siteSnapshotName)
		if err != nil {
			fail(ExitExtFailed, err.Error())
		}
		writeJSON(map[string]interface{}{
			"version": "1.0", "success": true,
			"version_name": ver.Version, "path": ver.Path,
			"message": fmt.Sprintf("Snapshot '%s' created", ver.Version),
		})
	},
}

// site version history
var siteVersionHistoryCmd = &cobra.Command{
	Use:   "history <site>",
	Short: "Show version history (active lineage, excludes orphaned)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _ := config.Load()
		db.Initialize(cfg.DatabasePath)
		defer db.Close()

		siteRepo := db.NewSiteRepository()
		site, err := siteRepo.GetBySlug(args[0])
		if err != nil {
			fail(ExitNotFound, "site not found")
		}

		verRepo := db.NewSiteVersionRepository()
		versions, err := verRepo.GetActiveLineage(site.ID)
		if err != nil {
			fail(ExitExtFailed, err.Error())
		}

		entries := []map[string]interface{}{}
		for _, v := range versions {
			entries = append(entries, map[string]interface{}{
				"version": v.Version, "is_active": v.IsActive,
				"comment": v.Comment, "created_at": v.CreatedAt,
			})
		}
		writeJSON(map[string]interface{}{"version": "1.0", "versions": entries})
	},
}

// site version prune
var siteVersionPruneCmd = &cobra.Command{
	Use:   "prune <site>",
	Short: "Delete all orphaned (dead-branch) versions",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, _ := config.Load()
		db.Initialize(cfg.DatabasePath)
		defer db.Close()

		svc := services.NewVersioningService(cfg)
		count, err := svc.PruneOrphaned(args[0])
		if err != nil {
			fail(ExitExtFailed, err.Error())
		}
		writeJSON(map[string]interface{}{
			"version": "1.0", "success": true,
			"deleted": count,
			"message": fmt.Sprintf("Pruned %d orphaned versions", count),
		})
	},
}


func init() {
	siteListCmd.Flags().String("target", "", "Remote target (host:port)")
	siteCreateCmd.Flags().String("name", "", "Site name")
	siteCreateCmd.Flags().String("slug", "", "Site slug")

	siteVersionCreateCmd.Flags().String("version", "", "Version (e.g., v1, v2)")
	siteVersionCreateCmd.Flags().String("comment", "", "Version comment")
	siteVersionCreateCmd.Flags().String("author", "", "Author name")
	siteVersionCreateCmd.Flags().String("target", "", "Remote target (host:port)")
	siteVersionListCmd.Flags().String("target", "", "Remote target (host:port)")
	siteVersionSwitchCmd.Flags().String("target", "", "Remote target (host:port)")
	siteWriteCmd.Flags().String("content", "", "File content")
	siteWriteCmd.Flags().String("target", "", "Remote target (host:port)")

	siteCmd.AddCommand(siteListCmd)
	siteCmd.AddCommand(siteCreateCmd)
	siteVersionSnapshotCmd.Flags().StringVar(&siteSnapshotName, "name", "", "Snapshot name (required)")
	siteVersionCmd.AddCommand(siteVersionSnapshotCmd)
	siteVersionCmd.AddCommand(siteVersionHistoryCmd)
	siteVersionCmd.AddCommand(siteVersionPruneCmd)
	siteVersionCmd.AddCommand(siteVersionCreateCmd)
	siteVersionCmd.AddCommand(siteVersionListCmd)
	siteVersionCmd.AddCommand(siteVersionSwitchCmd)
	siteCmd.AddCommand(siteVersionCmd)
	siteCmd.AddCommand(siteWriteCmd)
	siteCmd.AddCommand(siteDnsCmd)
	siteCmd.AddCommand(siteAdminCmd)
}
