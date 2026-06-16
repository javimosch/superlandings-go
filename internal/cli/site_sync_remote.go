package cli

import (
	"fmt"
	"os"
)

func handleRemoteSiteSync(target, siteSlug string) {
	client, err := NewRemoteClientFromTarget(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to target: %v\n", err)
		os.Exit(1)
	}
	
	// Trigger sync on remote daemon
	payload := map[string]interface{}{
		"site_slug": siteSlug,
	}
	
	result, err := client.SyncSite(siteSlug, payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error syncing remote site: %v\n", err)
		os.Exit(1)
	}
	
	if success, ok := result["success"].(bool); ok && success {
		fmt.Printf("Site synced successfully to remote target\n")
		if message, ok := result["message"].(string); ok {
			fmt.Printf("Message: %s\n", message)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Sync failed on remote target\n")
		os.Exit(1)
	}
}