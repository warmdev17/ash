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
	forceDelete      bool
	localForceDelete bool
)

var groupDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a top-level GitLab group",
	Long: `Delete a top-level GitLab group.
If run inside a simplified group folder (with .ash/group.json), it attempts to delete the current group.
Otherwise, a group name argument is required.

Safety: By default, it refuses to delete non-empty groups on GitLab.
Use --force (-f) to force deletion on GitLab.
Use --local-force (-l) to also delete the local folder.`,
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		var groupName string

		// 1. Context Awareness
		if len(args) > 0 {
			groupName = args[0]
		} else {
			// Check if we are in a group root
			wd, _ := os.Getwd()
			metaPath := filepath.Join(wd, ".ash", "group.json")
			if fileExists(metaPath) {
				var meta rootGroupMeta
				if err := readJSON(metaPath, &meta); err == nil && meta.Group.Name != "" {
					groupName = meta.Group.Name
					fmt.Printf("Detected current group: %s\n", groupName)
				}
			}
		}

		if groupName == "" {
			return fmt.Errorf("missing group name and not in a group folder")
		}

		return deleteGroup(groupName)
	},
}

func init() {
	groupCmd.AddCommand(groupDeleteCmd)
	groupDeleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Force delete on GitLab even if not empty")
	groupDeleteCmd.Flags().BoolVarP(&localForceDelete, "local-force", "l", false, "Also delete local folder")
}

func deleteGroup(name string) error {
	cfg, cfgPath, err := loadConfig()
	if err != nil {
		return err
	}

	g, ok := findGroupByName(cfg, name)
	if !ok {
		return fmt.Errorf("group %q not found in config (%s)", name, cfgPath)
	}

	fmt.Printf("Deleting group %q (ID: %d) from GitLab...\n", g.Name, g.ID)

	// API call logic: glab api -X DELETE /groups/:id
	// If force is not set, we trust GitLab API to fail if not empty (or we assume standard behavior).
	// But usually DELETE /groups/:id happens async or requires more permissions.
	// NOTE: GitLab API doesn't strictly block "non-empty" delete, it schedules for deletion.
	// But for safety locally, maybe we should warn?
	// The user asked "chi xoa khi trong".
	// We can check projects/subgroups count first.

	if !forceDelete {
		// Check emptiness
		sgs, _ := apiListSubgroups(g.ID)
		prjs, _ := apiListProjects(g.ID)
		if len(sgs) > 0 || len(prjs) > 0 {
			return fmt.Errorf("group is not empty (contains %d subgroups, %d projects). Use -f to force delete", len(sgs), len(prjs))
		}
	}

	// API Delete
	err = RunSpinner(fmt.Sprintf("Deleting group %s (ID: %d)", g.Name, g.ID), func() error {
		glabCmd := exec.Command("glab", "api", "-X", "DELETE", "/groups/"+strconv.FormatInt(g.ID, 10))
		if out, err := glabCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to delete group on GitLab: %s (%w)", string(out), err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Remove from config
	newGroups := []GitLabGroup{}
	for _, grp := range cfg.Groups {
		if grp.ID != g.ID {
			newGroups = append(newGroups, grp)
		}
	}
	cfg.Groups = newGroups
	if err := saveConfig(cfgPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	fmt.Printf("Updated config: %s\n", cfgPath)

	// Local delete
	if localForceDelete {
		// Caution: We need to know WHERE the local folder is.
		// If the user is running by Name and the folder is elsewhere, we might not know.
		// BUT if we assume standard structure or check common paths...
		// Best effort: if we are IN the folder, we can't fully remove it (busy).
		// If we are outside, we can look for it.
		// For now, let's assume we delete `slugify(name)` or the matched path in current dir?
		// Since we don't store local path in config, we can only guess 'Name' or 'Path' relative to CWD?
		// Or we only support this if context-aware?
		wd, _ := os.Getwd()
		target := filepath.Join(wd, g.Name)
		if !fileExists(target) {
			target = filepath.Join(wd, g.Path)
		}

		if fileExists(target) {
			if err := os.RemoveAll(target); err != nil {
				fmt.Printf("%s[WARN] Failed to delete local folder %s: %v%s\n", Yellow, target, err, Reset)
			} else {
				fmt.Printf("%s[DEL] Deleted local folder: %s%s\n", Red, target, Reset)
			}
		} else {
			// Handle case: we are inside the folder
			cwdName := filepath.Base(wd)
			if cwdName == g.Name || cwdName == g.Path {
				fmt.Printf("%s[WARN] Cannot delete local folder while inside it. Please cd .. and run again.%s\n", Yellow, Reset)
			}
		}
	}

	return nil
}
