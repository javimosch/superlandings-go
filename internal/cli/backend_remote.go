package cli

import (
	"fmt"
	"os"
)

func handleRemoteBackendStatus(target string) {
	// Parse target as host:port
	var host string
	var port int
	fmt.Sscanf(target, "%s:%d", &host, &port)
	
	if port == 0 {
		port = 3100 // default sl-cli daemon port
	}
	
	client := NewRemoteClient(host, port)
	result, err := client.GetStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking remote daemon status: %v\n", err)
		os.Exit(1)
	}
	
	status, ok := result["status"].(string)
	if !ok {
		fmt.Fprintf(os.Stderr, "Invalid response from remote daemon\n")
		os.Exit(1)
	}
	
	if status == "running" {
		fmt.Println("Daemon running")
		if service, ok := result["service"].(string); ok {
			fmt.Printf("Service: %s\n", service)
		}
	} else {
		fmt.Println("Daemon not running")
	}
}