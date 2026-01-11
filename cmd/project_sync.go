package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var projectSyncCmd = &cobra.Command{
	Use:   "sync [name...]",
	Short: "Sync (git pull) project code",
	Long: `Sync code for projects.
If no arguments provided: Syncs ALL projects in the current subgroup.
If arguments provided: Syncs only the specified projects.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, _ := os.Getwd()

		// Check context
		inSubgroup := fileExists(filepath.Join(wd, ".ash", "subgroup.json"))

		var targets []string // Folders to sync

		if len(args) > 0 {
			targets = args
		} else {
			if inSubgroup {
				// Scan all folders in subgroup that look like git repos
				entries, _ := os.ReadDir(wd)
				for _, e := range entries {
					if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
						if fileExists(filepath.Join(wd, e.Name(), ".git")) {
							targets = append(targets, e.Name())
						}
					}
				}
			} else {
				// Maybe we are INSIDE a project?
				if fileExists(filepath.Join(wd, ".git")) {
					// Sync current
					targets = []string{"."}
				} else {
					return fmt.Errorf("not in a subgroup or project folder")
				}
			}
		}

		if len(targets) == 0 {
			fmt.Println("No projects to sync.")
			return nil
		}

		fmt.Printf("Syncing %d projects (git pull)...\n", len(targets))

		// Concurrent Pull
		var wg sync.WaitGroup
		sem := make(chan struct{}, 5)

		for _, t := range targets {
			wg.Add(1)
			go func(dirName string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				targetDir := dirName
				if dirName != "." {
					targetDir = filepath.Join(wd, dirName)
				}

				if !fileExists(filepath.Join(targetDir, ".git")) {
					fmt.Printf("%s[SKIP] %s is not a git repo%s\n", Yellow, dirName, Reset)
					return
				}

				// Pull
				// Remove --quiet to see if "Already up to date" or updates
				out, err := exec.Command("git", "-C", targetDir, "pull").CombinedOutput()
				output := string(out)
				if err != nil {
					fmt.Printf("%s[ERR] %s pull failed: %v\n%s%s\n", Red, dirName, err, output, Reset)
				} else {
					if strings.Contains(output, "Already up to date") {
						// Clean check, maybe don't print anything or print in gray?
						// User wants "notification if specific action happened".
						fmt.Printf("%s[OK] %s is up to date%s\n", Gray, dirName, Reset)
					} else {
						// It updated
						fmt.Printf("%s[UPD] %s updated%s\n", Green, dirName, Reset)
					}
				}
			}(t)
		}
		wg.Wait()
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectSyncCmd)
}
