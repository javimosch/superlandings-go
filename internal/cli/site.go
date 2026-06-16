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
		target, _ := cmd.Flags().GetString("target")
		
		if target != "" {
			// Remote execution
			handleRemoteSiteList(target)
			return
		}
		
		// Local execution
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

		fmt.Println("ID\tName\tSlug\tDomains")
		dnsService := services.NewDNSService(cfg)
		for _, site := range sites {
			domains, err := dnsService.GetDomains(site.ID)
			domainStr := ""
			if err == nil && len(domains) > 0 {
				domainList := make([]string, len(domains))
				for i, d := range domains {
					domainList[i] = d.Domain
				}
				domainStr = fmt.Sprintf("%v", domainList)
			}
			fmt.Printf("%s\t%s\t%s\t%s\n", site.ID, site.Name, site.Slug, domainStr)
		}
	},
}

func init() {
	siteListCmd.Flags().String("target", "", "Remote target (host:port)")
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

// site dns
var siteDnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "Manage DNS via hotify-cli",
}

// site dns setup
var siteDnsSetupCmd = &cobra.Command{
	Use:   "setup <site>",
	Short: "Setup DNS for a site via hotify-cli",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		domain, _ := cmd.Flags().GetString("domain")
		ip, _ := cmd.Flags().GetString("ip")
		traefik, _ := cmd.Flags().GetBool("traefik")

		if domain == "" {
			fmt.Fprintf(os.Stderr, "Error: --domain is required\n")
			os.Exit(1)
		}
		if ip == "" {
			fmt.Fprintf(os.Stderr, "Error: --ip is required\n")
			os.Exit(1)
		}

		// Get site by slug
		siteService := services.NewSiteService(cfg)
		site, err := siteService.GetBySlug(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting site: %v\n", err)
			os.Exit(1)
		}

		// Setup DNS via hotify-cli
		dnsService := services.NewDNSService(cfg)
		if err := dnsService.SetupDNS(site.ID, site.Slug, domain, ip, traefik); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting up DNS: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("DNS setup successfully for %s -> %s\n", domain, ip)
		if traefik {
			fmt.Println("Traefik routing configured")
		}
	},
}

// site dns list
var siteDnsListCmd = &cobra.Command{
	Use:   "list <site>",
	Short: "List DNS domains for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Get site by slug
		siteService := services.NewSiteService(cfg)
		site, err := siteService.GetBySlug(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting site: %v\n", err)
			os.Exit(1)
		}

		// Get domains
		dnsService := services.NewDNSService(cfg)
		domains, err := dnsService.GetDomains(site.ID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting domains: %v\n", err)
			os.Exit(1)
		}

		if len(domains) == 0 {
			fmt.Println("No domains configured")
			return
		}

		fmt.Println("Domain\tIP\tTraefik")
		for _, d := range domains {
			traefik := "No"
			if d.Traefik {
				traefik = "Yes"
			}
			fmt.Printf("%s\t%s\t%s\n", d.Domain, d.IP, traefik)
		}
	},
}

// site dns remove
var siteDnsRemoveCmd = &cobra.Command{
	Use:   "remove <site>",
	Short: "Remove DNS configuration for a site",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeDB(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing database: %v\n", err)
			os.Exit(1)
		}
		defer db.Close()

		// Remove DNS via hotify-cli
		dnsService := services.NewDNSService(cfg)
		if err := dnsService.RemoveDNS(args[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing DNS: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("DNS configuration removed for %s\n", args[0])
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

	// DNS commands
	siteDnsSetupCmd.Flags().String("domain", "", "Domain name (e.g., slv2.intrane.fr)")
	siteDnsSetupCmd.Flags().String("ip", "", "IP address (e.g., 92.113.145.16)")
	siteDnsSetupCmd.Flags().Bool("traefik", false, "Setup Traefik routing")

	siteDnsCmd.AddCommand(siteDnsSetupCmd)
	siteDnsCmd.AddCommand(siteDnsListCmd)
	siteDnsCmd.AddCommand(siteDnsRemoveCmd)

	siteCmd.AddCommand(siteVersionCmd)
	siteCmd.AddCommand(siteWriteCmd)
	siteCmd.AddCommand(siteDnsCmd)
}