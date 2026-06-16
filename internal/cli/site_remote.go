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
	
	fmt.Println("Slug\tName")
	for _, s := range sites {
		site := s.(map[string]interface{})
		fmt.Printf("%s\t%s\n", site["slug"], site["name"])
	}
}