package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/spf13/cobra"
)

var projectSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync all projects (Clone/Pull)",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		subMetaPath := filepath.Join(wd, ".ash", "subgroup.json")
		if !fileExists(subMetaPath) {
			return fmt.Errorf("not in a subgroup folder")
		}

		var meta subgroupMeta
		if err := readJSON(subMetaPath, &meta); err != nil {
			return err
		}

		if len(meta.Projects) == 0 {
			fmt.Printf("%s[INFO] No projects found in metadata.%s\n", Cyan, Reset)
			return nil
		}

		// EXECUTE
		var results []TaskResult
		var mu sync.Mutex

		title := fmt.Sprintf("Syncing %d project(s)...", len(meta.Projects))

		err = RunWithSpinner(title, func() error {
			var wg sync.WaitGroup
			sem := make(chan struct{}, 5) // Limit 5 concurrent threads

			for _, p := range meta.Projects {
				wg.Add(1)
				go func(proj projectIdent) {
					defer wg.Done()
					sem <- struct{}{}
					defer func() { <-sem }()

					res := syncOneProject(wd, proj)

					mu.Lock()
					results = append(results, res)
					mu.Unlock()
				}(p)
			}
			wg.Wait()
			return nil
		})
		if err != nil {
			return err
		}
		PrintResults(results)
		return nil
	},
}

func syncOneProject(wd string, p projectIdent) TaskResult {
	targetDir := filepath.Join(wd, p.Name)

	// Case 1: Folder missing -> CLONE
	if !fileExists(targetDir) {
		// Dùng glab repo clone để nó tự handle auth/url
		err := exec.Command("glab", "repo", "clone", p.Path, targetDir).Run()
		if err != nil {
			return TaskResult{Name: p.Name, Status: "ERR", Message: "Clone failed"}
		}
		return TaskResult{Name: p.Name, Status: "NEW", Message: "Cloned"}
	}

	// Case 2: Folder exists -> PULL
	if !fileExists(filepath.Join(targetDir, ".git")) {
		return TaskResult{Name: p.Name, Status: "SKIP", Message: "Not a git repo"}
	}

	err := exec.Command("git", "-C", targetDir, "pull", "--quiet").Run()
	if err != nil {
		return TaskResult{Name: p.Name, Status: "ERR", Message: "Pull failed (Conflict?)"}
	}
	return TaskResult{Name: p.Name, Status: "OK", Message: "Updated"}
}

func init() {
	projectCmd.AddCommand(projectSyncCmd)
}
