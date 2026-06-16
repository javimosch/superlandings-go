package cli

import (
	"fmt"
	"os"
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