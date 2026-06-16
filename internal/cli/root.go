package cli

import (
	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/spf13/cobra"
)

var cfg *config.Config

func Execute(config *config.Config) error {
	cfg = config
	return rootCmd.Execute()
}

var rootCmd = &cobra.Command{
	Use:   "sl-cli",
	Short: "SuperLandings CLI - Manage landing pages",
	Long:  `SuperLandings CLI is a tool for managing landing pages with support for multiple types including HTML, EJS, virtual, and static sites.`,
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(landingCmd)
	rootCmd.AddCommand(backendCmd)
	rootCmd.AddCommand(organizationCmd)
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(systemdCmd)
	rootCmd.AddCommand(siteCmd)
}