package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	withProjects bool
	cleanLocal   bool
)

var subgroupSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync subgroup metadata",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		ashDir := filepath.Join(wd, ".ash")
		groupMetaPath := filepath.Join(ashDir, "group.json")

		if !fileExists(groupMetaPath) {
			return fmt.Errorf("not in a group root (.ash/group.json missing)")
		}

		var meta rootGroupMeta
		if err := readJSON(groupMetaPath, &meta); err != nil {
			return err
		}

		var newSubs []subgroupIdent
		var removed []string

		// EXECUTE
		err = RunWithSpinner("Fetching subgroups from GitLab...", func() error {
			// 1. API Call (Manual exec to reuse helper if needed, or inline)
			url := fmt.Sprintf("groups/%d/subgroups?per_page=100", meta.Group.ID)
			out, err := exec.Command("glab", "api", url, "--paginate").Output()
			if err != nil {
				return err
			}

			var sgs []glGroup
			if err := json.Unmarshal(out, &sgs); err != nil {
				return err
			}

			// 2. Diff Logic
			newNameSet := make(map[string]bool)
			for _, sg := range sgs {
				newSubs = append(newSubs, subgroupIdent{ID: sg.ID, Path: sg.Path, Name: sg.Name})
				newNameSet[strings.ToLower(sg.Name)] = true
			}

			for _, old := range meta.Subgroups {
				if !newNameSet[strings.ToLower(old.Name)] {
					removed = append(removed, old.Name)
				}
			}

			// 3. Write
			meta.Subgroups = newSubs
			return writeJSON(groupMetaPath, meta)
		})
		if err != nil {
			return err
		}

		// Report
		fmt.Printf("%s[OK] Synced %d subgroups.%s\n", Green, len(newSubs), Reset)
		if len(removed) > 0 {
			fmt.Printf("%s[INFO] Removed from GitLab: %v%s\n", Yellow, removed, Reset)
			if cleanLocal {
				for _, r := range removed {
					os.RemoveAll(filepath.Join(wd, r))
					fmt.Printf("%s[DEL] Deleted folder: %s%s\n", Red, r, Reset)
				}
			}
		}

		return nil
	},
}

func init() {
	subgroupCmd.AddCommand(subgroupSyncCmd)
	subgroupSyncCmd.Flags().BoolVar(&withProjects, "with-projects", false, "Sync projects metadata too")
	subgroupSyncCmd.Flags().BoolVar(&cleanLocal, "clean", false, "Remove local folders of deleted subgroups")
}
