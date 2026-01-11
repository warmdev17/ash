package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var sgSyncClean bool

var subgroupSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync projects in the current subgroup",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		subMetaPath := filepath.Join(wd, ".ash", "subgroup.json")

		if !fileExists(subMetaPath) {
			return fmt.Errorf("not in a subgroup folder (.ash/subgroup.json missing)")
		}

		var meta subgroupMeta
		if err := readJSON(subMetaPath, &meta); err != nil {
			return err
		}

		fmt.Printf("Syncing Subgroup: %s (ID: %d)\n", meta.Group.Name, meta.Group.ID)

		// Use the shared helper from group_sync.go
		if err := syncSubgroupContent(wd, meta.Group.ID, sgSyncClean); err != nil {
			return err
		}

		fmt.Printf("%s[OK] Sync complete.%s\n", Green, Reset)
		return nil
	},
}

func init() {
	subgroupCmd.AddCommand(subgroupSyncCmd)
	subgroupSyncCmd.Flags().BoolVar(&sgSyncClean, "clean", false, "Delete local folders of removed projects")
}
