package cli

import (
	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/javimosch/superlandings-go/internal/db"
	"github.com/spf13/cobra"
)

var cfg *config.Config

func Execute(config *config.Config) error {
	cfg = config
	return rootCmd.Execute()
}

func initializeDB() error {
	if err := cfg.EnsureDirectories(); err != nil {
		return err
	}
	return db.Initialize(cfg.DatabasePath)
}

var rootCmd = &cobra.Command{
	Use:   "sl-cli",
	Short: "SuperLandings CLI - Manage static sites with versioning",
	Long:  `SuperLandings CLI is a tool for managing static sites with version control, Go templates, assets, blog, forms, and domain-aware serving.`,
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(backendCmd)
	rootCmd.AddCommand(userCmd)
	rootCmd.AddCommand(systemdCmd)
	rootCmd.AddCommand(siteCmd)
	rootCmd.AddCommand(targetsCmd)
}