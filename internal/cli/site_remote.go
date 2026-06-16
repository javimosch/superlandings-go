package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func handleRemoteSiteList(target string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to target: %v\n", err)
		os.Exit(1)
	}
	
	result, err := client.ListSites()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing remote sites: %v\n", err)
		os.Exit(1)
	}
	
	// Parse JSON response (parseResponse wraps arrays in {"sites": [...]})
	sites, ok := result["sites"].([]interface{})
	if !ok {
		fmt.Fprintf(os.Stderr, "Invalid response from remote daemon\n")
		os.Exit(1)
	}
	
	if len(sites) == 0 {
		fmt.Println("No sites found")
		return
	}
	
	fmt.Println("ID\tName\tSlug\tDomains")
	for _, s := range sites {
		site := s.(map[string]interface{})
		id := ""
		if idVal, ok := site["id"].(string); ok {
			id = idVal
		}
		name := site["name"].(string)
		slug := site["slug"].(string)
		domains := ""
		if domainsVal, ok := site["domains"].([]interface{}); ok && len(domainsVal) > 0 {
			domainList := make([]string, len(domainsVal))
			for i, d := range domainsVal {
				domainList[i] = d.(string)
			}
			domains = fmt.Sprintf("%v", domainList)
		}
		fmt.Printf("%s\t%s\t%s\t%s\n", id, name, slug, domains)
	}
}

func handleRemoteVersionList(target string, siteSlug string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to target: %v\n", err)
		os.Exit(1)
	}
	
	result, err := client.ListVersions(siteSlug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing remote versions: %v\n", err)
		os.Exit(1)
	}
	
	// Parse JSON response
	versions, ok := result["versions"].([]interface{})
	if !ok {
		fmt.Fprintf(os.Stderr, "Invalid response from remote daemon\n")
		os.Exit(1)
	}
	
	if len(versions) == 0 {
		fmt.Println("No versions found")
		return
	}
	
	fmt.Println("Version\tPath\tActive\tComment")
	for _, v := range versions {
		version := v.(map[string]interface{})
		versionStr := version["version"].(string)
		path := version["path"].(string)
		active := "No"
		if activeVal, ok := version["is_active"].(bool); ok && activeVal {
			active = "Yes"
		}
		comment := ""
		if commentVal, ok := version["comment"].(string); ok {
			comment = commentVal
		}
		fmt.Printf("%s\t%s\t%s\t%s\n", versionStr, path, active, comment)
	}
}

func handleRemoteVersionCreate(target string, args []string, cmd *cobra.Command) {
	version, _ := cmd.Flags().GetString("version")
	comment, _ := cmd.Flags().GetString("comment")
	author, _ := cmd.Flags().GetString("author")
	
	if version == "" {
		fmt.Fprintf(os.Stderr, "Error: --version is required\n")
		os.Exit(1)
	}
	
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to target: %v\n", err)
		os.Exit(1)
	}
	
	result, err := client.CreateVersion(args[0], version, comment, author)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating remote version: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Version created successfully!")
	if versionVal, ok := result["version"].(string); ok {
		fmt.Printf("Version: %s\n", versionVal)
	}
	if pathVal, ok := result["path"].(string); ok {
		fmt.Printf("Path: %s\n", pathVal)
	}
	if isActive, ok := result["is_active"].(bool); ok && isActive {
		fmt.Println("This version is now active")
	}
}

func handleRemoteVersionSwitch(target string, siteSlug, version string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to target: %v\n", err)
		os.Exit(1)
	}
	
	_, err = client.SwitchVersion(siteSlug, version)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error switching remote version: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Switched to version %s\n", version)
}

func handleRemoteSiteWrite(target string, args []string, cmd *cobra.Command) {
	content, _ := cmd.Flags().GetString("content")
	
	if content == "" {
		fmt.Fprintf(os.Stderr, "Error: --content is required\n")
		os.Exit(1)
	}
	
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to target: %v\n", err)
		os.Exit(1)
	}
	
	_, err = client.WriteFile(args[0], args[1], args[2], content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing remote file: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("File written successfully: %s\n", args[2])
}

func handleRemoteDNSList(target string, siteSlug string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to target: %v\n", err)
		os.Exit(1)
	}
	
	result, err := client.ListDNS(siteSlug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing remote DNS: %v\n", err)
		os.Exit(1)
	}
	
	// Parse JSON response
	domains, ok := result["domains"].([]interface{})
	if !ok {
		fmt.Fprintf(os.Stderr, "Invalid response from remote daemon\n")
		os.Exit(1)
	}
	
	if len(domains) == 0 {
		fmt.Println("No DNS entries found")
		return
	}
	
	fmt.Println("Domain\tIP\tTraefik")
	for _, d := range domains {
		domain := d.(map[string]interface{})
		domainStr := domain["domain"].(string)
		ip := domain["ip"].(string)
		traefik := "No"
		if traefikVal, ok := domain["traefik"].(bool); ok && traefikVal {
			traefik = "Yes"
		}
		fmt.Printf("%s\t%s\t%s\n", domainStr, ip, traefik)
	}
}

func handleRemoteDNSSetup(target string, args []string, cmd *cobra.Command) {
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
	
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to target: %v\n", err)
		os.Exit(1)
	}
	
	_, err = client.SetupDNS(args[0], domain, ip, traefik)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up remote DNS: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("DNS setup successfully for %s -> %s\n", domain, ip)
	if traefik {
		fmt.Println("Traefik routing configured")
	}
}

func handleRemoteDNSRemove(target string, args []string) {
	// args[0] is site slug
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to target: %v\n", err)
		os.Exit(1)
	}
	
	_, err = client.RemoveDNS(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error removing remote DNS: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("DNS configuration removed for site %s\n", args[0])
}