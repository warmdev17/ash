package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	withProjects bool // --with-projects: also write .ash/subgroup.json for each subgroup (no clone)
	cleanLocal   bool // --clean: remove local subgroup directories that no longer exist on GitLab
)

var subgroupSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync subgroup metadata from GitLab (updates .ash/group.json)",
	Example: `  cd MyGroup
  ash subgroup sync
  ash subgroup sync --with-projects
  ash subgroup sync --clean
  ash subgroup sync --clean --with-projects`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Must be in group root
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd failed: %w", err)
		}
		ashDir := filepath.Join(wd, ".ash")
		groupMetaPath := filepath.Join(ashDir, "group.json")
		subMetaPath := filepath.Join(ashDir, "subgroup.json")

		if !fileExists(groupMetaPath) {
			return fmt.Errorf("not in a group root: %s not found", groupMetaPath)
		}
		if fileExists(subMetaPath) {
			return fmt.Errorf("this looks like a subgroup folder (found %s); run from the group root", subMetaPath)
		}

		// Read current group meta (to get group ID) and keep a copy of old subgroups
		var meta rootGroupMeta
		if err := readJSON(groupMetaPath, &meta); err != nil {
			return fmt.Errorf("parse group.json failed: %w", err)
		}
		if meta.Group.ID == 0 {
			return fmt.Errorf("invalid group.json: missing group.id")
		}
		oldSubs := append([]subgroupIdent(nil), meta.Subgroups...) // copy

		// Fetch latest subgroups from GitLab
		sgs, err := apiListSubgroups(meta.Group.ID)
		if err != nil {
			return err
		}

		// Build new Subgroups list
		newSubs := make([]subgroupIdent, 0, len(sgs))
		newNameSet := make(map[string]struct{}, len(sgs)) // lowercased name set for diff
		for _, sg := range sgs {
			newSubs = append(newSubs, subgroupIdent{
				ID:   sg.ID,
				Path: sg.Path,
				Name: sg.Name,
			})
			newNameSet[strings.ToLower(sg.Name)] = struct{}{}
		}

		// Detect removed subgroups by name (case-insensitive)
		removed := make([]string, 0)
		for _, old := range oldSubs {
			if _, ok := newNameSet[strings.ToLower(old.Name)]; !ok {
				removed = append(removed, old.Name)
			}
		}

		// Optionally clean local dirs for removed subgroups
		if cleanLocal && len(removed) > 0 {
			fmt.Println("Cleaning local subgroup directories that no longer exist on GitLab...")
			for _, name := range removed {
				sgDir := filepath.Join(wd, name)
				if _, err := os.Stat(sgDir); err == nil {
					if err := os.RemoveAll(sgDir); err != nil {
						fmt.Printf("Failed to remove %s: %v\n", sgDir, err)
					} else {
						fmt.Printf("Removed: %s\n", sgDir)
					}
				}
			}
			fmt.Println("Cleanup complete.")
		}

		// Overwrite .ash/group.json -> Subgroups (always prune removed in metadata)
		meta.Subgroups = newSubs
		if err := writeGroupJSON(ashDir, meta); err != nil {
			return fmt.Errorf("update group.json failed: %w", err)
		}
		fmt.Printf("Synced %d subgroups to %s\n", len(newSubs), groupMetaPath)
		if len(removed) > 0 {
			fmt.Printf("Pruned %d removed subgroup(s) from metadata: %v\n", len(removed), removed)
			if !cleanLocal {
				fmt.Println("Note: local folders were kept (use --clean to remove them).")
			}
		}

		// Optionally, also write per-subgroup metadata (no cloning)
		if withProjects {
			for _, sg := range newSubs {
				// Create folder if missing (named by Display Name for readability)
				sgDir := filepath.Join(wd, sg.Name)
				if err := os.MkdirAll(sgDir, 0o755); err != nil {
					return fmt.Errorf("create subgroup dir %q: %w", sgDir, err)
				}

				// Fetch projects in this subgroup and write .ash/subgroup.json
				prjs, err := apiListProjects(sg.ID)
				if err != nil {
					return fmt.Errorf("list projects for subgroup %q failed: %w", sg.Name, err)
				}
				prIdents := make([]projectIdent, 0, len(prjs))
				for _, p := range prjs {
					prIdents = append(prIdents, projectIdent{ID: p.ID, Path: p.Path, Name: p.Name})
				}
				if err := writeSubgroupJSON(filepath.Join(sgDir, ".ash"), subgroupMeta{
					Group:    groupIdent{ID: sg.ID, Path: sg.Path},
					Projects: prIdents,
				}); err != nil {
					return fmt.Errorf("write %s failed: %w", filepath.Join(sgDir, ".ash", "subgroup.json"), err)
				}
				fmt.Printf("Wrote %s (.ash/subgroup.json) with %d project(s)\n", sgDir, len(prIdents))
			}
		}

		fmt.Println("Done.")
		return nil
	},
}

func init() {
	subgroupCmd.AddCommand(subgroupSyncCmd)
	subgroupSyncCmd.Flags().BoolVar(&withProjects, "with-projects", false, "Also write .ash/subgroup.json for each subgroup (no clone)")
	subgroupSyncCmd.Flags().BoolVar(&cleanLocal, "clean", false, "Remove local subgroup directories that no longer exist on GitLab")
}
