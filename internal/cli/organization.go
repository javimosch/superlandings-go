package cli

import (
	"github.com/spf13/cobra"
)

var organizationCmd = &cobra.Command{
	Use:   "organization",
	Short: "Manage organizations",
	Long:  `Create, read, update, and delete organizations for landing page management.`,
}

// organization list
var organizationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all organizations",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement organization list
		println("Organization list - TODO")
	},
}

// organization create
var organizationCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new organization",
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: Implement organization create
		println("Organization create - TODO")
	},
}

func init() {
	organizationCmd.AddCommand(organizationListCmd)
	organizationCmd.AddCommand(organizationCreateCmd)
}