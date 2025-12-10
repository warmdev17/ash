package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var projectSyncCmd = &cobra.Command{
	Use: "sync",

	Short: "Sync projects between GitLab and local workspace",
	Long: `Synchronize all projects in the current subgroup between GitLab and the local workspace.

New projects on GitLab will be cloned automatically.
Deleted projects on GitLab will be removed from subgroup.json.

Use the --clean flag to also delete local folders corresponding
to projects that no longer exist on GitLab.

Examples:
  ash project sync
    → Updates metadata only (keeps local folders)
  ash project sync --clean
    → Deletes local folders for removed projects and updates metadata.`,
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd failed: %w", err)
		}
		ashDir := filepath.Join(wd, ".ash")
		// groupMetaPath := filepath.Join(ashDir, "groups.json")
		subMetaPath := filepath.Join(ashDir, "subgroups.json")

		if !fileExists(subMetaPath) {
			return fmt.Errorf(RED+"this look like a group folder (found %s); run from the subgroup folder"+RED, subMetaPath)
		}

		println(GREEN + "Sync called" + GREEN)

		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectSyncCmd)
}
