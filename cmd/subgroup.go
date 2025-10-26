package cmd

import "github.com/spf13/cobra"

var subgroupCmd = &cobra.Command{
	Use:   "subgroup",
	Short: "Manage subgroups under the current group",
	Long:  "Subgroup-level operations: create a subgroup or sync subgroups metadata from GitLab.",
}

func init() {
	rootCmd.AddCommand(subgroupCmd)
}
