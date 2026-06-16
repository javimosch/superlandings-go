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

func handleRemoteVersionCreate(target string, args []string) {
	fmt.Fprintf(os.Stderr, "Remote version creation not yet implemented\n")
	fmt.Fprintf(os.Stderr, "Use SSH directly: ssh user@server 'sl-cli site version create %s'\n", args[0])
	os.Exit(1)
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