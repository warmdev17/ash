package cmd

import (
	"github.com/spf13/cobra"
)

var projectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects within a subgroup",
	Long:  "Handle project-level operations in the current subgroup, including creating new projects and synchronizing with GitLab. Use 'ash project create' to generate multiple projects automatically, or 'ash project sync' to keep your local workspace in sync with GitLab.",
}

func init() {
	rootCmd.AddCommand(projectCmd)
}
