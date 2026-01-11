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
	prjForceDelete      bool
	prjLocalForceDelete bool
)

var projectDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a project (on GitLab and local)",
	Args:  cobra.MaximumNArgs(1), // 0 args if inside project, 1 arg if outside
	RunE: func(cmd *cobra.Command, args []string) error {
		var name string
		wd, _ := os.Getwd()

		// 1. Determine Context & Name
		inSubgroup := fileExists(filepath.Join(wd, ".ash", "subgroup.json"))

		if len(args) > 0 {
			name = args[0]
		} else {
			// Infer from current dir?
			// But wait, if we are INSIDE the project folder, typically there is no .ash file inside the project itself
			// (unless we add one, but we don't).
			// So we check if parent has subgroup.json?
			parent := filepath.Dir(wd)
			if fileExists(filepath.Join(parent, ".ash", "subgroup.json")) {
				name = filepath.Base(wd)
				// We need to jump up to parent to Execute delete (because we cant delete current dir easily)
				if prjLocalForceDelete {
					return fmt.Errorf("cannot delete local folder while inside it. Please cd .. and run 'ash project delete %s'", name)
				}
				// If just remote delete, it's fine.
			} else {
				return fmt.Errorf("missing project name and not inside a project folder")
			}
		}

		// 2. Load Meta (from current dir or parent)
		var metaPath string
		if inSubgroup {
			metaPath = filepath.Join(wd, ".ash", "subgroup.json")
		} else {
			// Try parent
			parent := filepath.Dir(wd)
			if fileExists(filepath.Join(parent, ".ash", "subgroup.json")) {
				metaPath = filepath.Join(parent, ".ash", "subgroup.json")
				wd = parent // Set WD to subgroup root for easier handling
			} else {
				return fmt.Errorf("must be run inside a subgroup or project folder")
			}
		}

		var meta subgroupMeta
		if err := readJSON(metaPath, &meta); err != nil {
			return err
		}

		// Find ID
		var targetID int64
		var targetPath string
		found := false
		for _, p := range meta.Projects {
			if p.Name == name {
				targetID = p.ID
				targetPath = p.Path
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("project %q not found in metadata", name)
		}

		fmt.Printf("Deleting project %s (ID: %d)...\n", name, targetID)

		// Check Empty? Projects are rarely "empty" in the sense of groups.
		// GitLab API deletes repo.
		// User said: "xoa khi trong".
		// Maybe check if repo has commits?
		// Hard to check without cloning.
		// Let's assume strict "-f" is required for ANY project delete to be safe,
		// OR we trust the "Empty" definition of GitLab (fresh repo).
		// But let's assume if it has Files, it's not empty.
		if !prjForceDelete {
			// We can list repository tree?
			// glab api projects/:id/repository/tree
			// If error or empty list -> Empty.
			cmd := exec.Command("glab", "api", fmt.Sprintf("/projects/%d/repository/tree", targetID))
			if out, _ := cmd.CombinedOutput(); len(out) > 5 { // json "[]" is 2 bytes
				return fmt.Errorf("project is not empty. Use -f to force")
			}
		}

		// API Delete
		// API Delete
		err := RunSpinner(fmt.Sprintf("Deleting project %s (ID: %d)", name, targetID), func() error {
			glabCmd := exec.Command("glab", "api", "-X", "DELETE", "/projects/"+strconv.FormatInt(targetID, 10))
			if out, err := glabCmd.CombinedOutput(); err != nil {
				return fmt.Errorf("gitlab delete failed: %s (%w)", string(out), err)
			}
			return nil
		})
		if err != nil {
			return err
		}

		// Update Meta
		newPrjs := []projectIdent{}
		for _, p := range meta.Projects {
			if p.ID != targetID {
				newPrjs = append(newPrjs, p)
			}
		}
		meta.Projects = newPrjs
		writeSubgroupJSON(metaPath, meta)

		// Local Delete
		if prjLocalForceDelete {
			localPath := filepath.Join(wd, name)
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
	projectCmd.AddCommand(projectDeleteCmd)
	projectDeleteCmd.Flags().BoolVarP(&prjForceDelete, "force", "f", false, "Force delete on GitLab")
	projectDeleteCmd.Flags().BoolVarP(&prjLocalForceDelete, "local-force", "l", false, "Delete local folder")
}
