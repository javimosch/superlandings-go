package cli

import (
	"fmt"
	"os"

	"github.com/javimosch/superlandings-go/internal/config"
	"github.com/spf13/cobra"
)

var targetsCmd = &cobra.Command{
	Use:   "targets",
	Short: "Manage remote targets",
	Long:  `Add, list, remove, and use remote targets for sl-cli operations.`,
}

var targetsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured targets",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadCLIConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		if len(cfg.Targets) == 0 {
			fmt.Println("No targets configured")
			return
		}

		fmt.Println("Name\tHost\tPort\tDefault")
		for _, target := range cfg.Targets {
			def := ""
			if target.Default {
				def = "*"
			}
			fmt.Printf("%s\t%s\t%d\t%s\n", target.Name, target.Host, target.Port, def)
		}
	},
}

var targetsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new target",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		host, _ := cmd.Flags().GetString("host")
		port, _ := cmd.Flags().GetInt("port")
		token, _ := cmd.Flags().GetString("token")
		setDefault, _ := cmd.Flags().GetBool("default")

		if name == "" || host == "" {
			fmt.Fprintf(os.Stderr, "Error: --name and --host are required\n")
			os.Exit(1)
		}

		if port == 0 {
			port = 3100 // default port
		}

		target := config.Target{
			Name:      name,
			Host:      host,
			Port:      port,
			AuthToken: token,
			Default:   setDefault,
		}

		if err := config.AddTarget(target); err != nil {
			fmt.Fprintf(os.Stderr, "Error adding target: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Target '%s' added successfully\n", name)
	},
}

var targetsRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a target",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprintf(os.Stderr, "Error: target name required\n")
			os.Exit(1)
		}

		name := args[0]

		if err := config.RemoveTarget(name); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing target: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Target '%s' removed successfully\n", name)
	},
}

func init() {
	targetsAddCmd.Flags().String("name", "", "Target name")
	targetsAddCmd.Flags().String("host", "", "Target host (IP or domain)")
	targetsAddCmd.Flags().Int("port", 3100, "Target port (default: 3100)")
	targetsAddCmd.Flags().String("token", "", "API authentication token")
	targetsAddCmd.Flags().Bool("default", false, "Set as default target")

	targetsCmd.AddCommand(targetsListCmd)
	targetsCmd.AddCommand(targetsAddCmd)
	targetsCmd.AddCommand(targetsRemoveCmd)
}