package cmd

import "github.com/spf13/cobra"

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage GitLab groups (create, clone, list, ...)",
	Long:  "Group-level operations: create new top-level groups or clone existing ones with full hierarchy.",
}

func init() {
	rootCmd.AddCommand(groupCmd)
}
