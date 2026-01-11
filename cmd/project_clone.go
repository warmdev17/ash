package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var projectCloneCmd = &cobra.Command{
	Use:   "clone [name]",
	Short: "Clone a specific project",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		wd, _ := os.Getwd()

		// Ensure we are in a subgroup for metadata update?
		// User might want to clone just to check, but our tool relies on structure.
		// Let's enforce structure for consistency.
		subMetaPath := filepath.Join(wd, ".ash", "subgroup.json")
		if !fileExists(subMetaPath) {
			return fmt.Errorf("not in a subgroup folder (.ash/subgroup.json missing)")
		}

		var meta subgroupMeta
		readJSON(subMetaPath, &meta)

		// 1. Find Remote Project
		// Search in local meta first? If it's missing locally but exists remotely.
		// Search API.
		found := false
		var target glProject

		prjs, err := apiListProjects(meta.Group.ID)
		if err != nil {
			return err
		}
		for _, p := range prjs {
			if p.Name == name {
				target = p
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("project %q not found on GitLab", name)
		}

		fmt.Printf("Cloning project %s...\n", target.Name)

		// 2. Clone
		dest := filepath.Join(wd, target.Name)
		if fileExists(dest) {
			return fmt.Errorf("folder %s already exists", dest)
		}

		err = RunSpinner(fmt.Sprintf("Cloning project %s", target.Name), func() error {
			if err := exec.Command("git", "clone", "--quiet", target.HTTPURLToRepo, dest).Run(); err != nil {
				return fmt.Errorf("git clone failed: %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}

		// 3. Update Meta
		exists := false
		for _, p := range meta.Projects {
			if p.ID == target.ID {
				exists = true
				break
			}
		}
		if !exists {
			meta.Projects = append(meta.Projects, projectIdent{ID: target.ID, Name: target.Name, Path: target.Path})
			writeSubgroupJSON(subMetaPath, meta)
		}

		fmt.Println("[OK] Project cloned.")
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectCloneCmd)
}
