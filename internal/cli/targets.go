package cli

import (
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
			fail(ExitExtFailed, err.Error())
		}
		if cfg.Targets == nil {
			cfg.Targets = []config.Target{}
		}
		writeJSON(map[string]interface{}{"version": "1.0", "targets": cfg.Targets})
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
			fail(ExitMissingFlag, "--name and --host are required")
		}
		if port == 0 {
			port = 3100
		}

		target := config.Target{Name: name, Host: host, Port: port, AuthToken: token, Default: setDefault}
		if err := config.AddTarget(target); err != nil {
			fail(ExitConflict, err.Error())
		}

		success("Target added successfully", map[string]interface{}{
			"name": name, "host": host, "port": port,
		})
	},
}

var targetsRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a target",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.RemoveTarget(args[0]); err != nil {
			fail(ExitNotFound, err.Error())
		}
		success("Target removed successfully", nil)
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
