package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	sgForceDelete      bool
	sgLocalForceDelete bool
)

var subgroupDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a subgroup (on GitLab and local)",
	Long: `Delete a subgroup.
Usage:
  ash subgroup delete Session1
  (Must be run from the Group root directory)

Behavior:
  - Deletes from GitLab.
  - Removes from parent group.json.
  - Optional: -l to delete local folder.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		wd, _ := os.Getwd()
		groupMetaPath := filepath.Join(wd, ".ash", "group.json")

		if !fileExists(groupMetaPath) {
			return fmt.Errorf("not in a group root")
		}

		var meta rootGroupMeta
		if err := readJSON(groupMetaPath, &meta); err != nil {
			return err
		}

		// Find ID
		var targetID int64
		var targetPath string
		found := false
		for _, sg := range meta.Subgroups {
			if sg.Name == name {
				targetID = sg.ID
				targetPath = sg.Path
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("subgroup %q not found in metadata", name)
		}

		fmt.Printf("Deleting subgroup %s (ID: %d)...\n", name, targetID)

		// Check Empty
		if !sgForceDelete {
			prjs, _ := apiListProjects(targetID)
			if len(prjs) > 0 {
				return fmt.Errorf("subgroup is not empty (%d projects). Use -f to force", len(prjs))
			}
		}

		// API Delete
		// API Delete
		err := RunSpinner(fmt.Sprintf("Deleting subgroup %s (ID: %d)", name, targetID), func() error {
			glabCmd := exec.Command("glab", "api", "-X", "DELETE", "/groups/"+strconv.FormatInt(targetID, 10))
			if out, err := glabCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("gitlab delete failed: %s (%w)", string(out), err)
			}
			return nil
		})
		if err != nil {
			return err
		}

		// Update Meta
		newSgs := []subgroupIdent{}
		for _, sg := range meta.Subgroups {
			if sg.ID != targetID {
				newSgs = append(newSgs, sg)
			}
		}
		meta.Subgroups = newSgs
		writeGroupJSON(filepath.Join(wd, ".ash"), meta)

		// Local Delete
		if sgLocalForceDelete {
			// Try deleting folder with same name
			localPath := filepath.Join(wd, name)
			// fallback if name != path? usually name matches folder in our flow.
			if !fileExists(localPath) && fileExists(filepath.Join(wd, targetPath)) {
				localPath = filepath.Join(wd, targetPath)
			}
			os.RemoveAll(localPath)
			fmt.Println("[DEL] Deleted local folder.")
		}

		return nil
	},
}

func init() {
	subgroupCmd.AddCommand(subgroupDeleteCmd)
	subgroupDeleteCmd.Flags().BoolVarP(&sgForceDelete, "force", "f", false, "Force delete on GitLab")
	subgroupDeleteCmd.Flags().BoolVarP(&sgLocalForceDelete, "local-force", "l", false, "Delete local folder")
}
