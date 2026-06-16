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
	fmt.Fprintf(os.Stderr, "Remote DNS list not yet implemented\n")
	fmt.Fprintf(os.Stderr, "Use SSH directly: ssh user@server 'sl-cli site dns list %s'\n", siteSlug)
	os.Exit(1)
}

func handleRemoteDNSSetup(target string, args []string, cmd *cobra.Command) {
	fmt.Fprintf(os.Stderr, "Remote DNS setup not yet implemented\n")
	fmt.Fprintf(os.Stderr, "Use SSH directly: ssh user@server 'sl-cli site dns setup %s'\n", args[0])
	os.Exit(1)
}