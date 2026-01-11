package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var (
	submitAll bool
	submitMsg string
)

var submitCmd = &cobra.Command{
	Use:   "submit [folder...]",
	Short: "Submit assignments",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		subMetaPath := filepath.Join(wd, ".ash", "subgroup.json")
		if !fileExists(subMetaPath) {
			return errors.New("not in a subgroup folder")
		}

		var meta subgroupMeta
		readJSON(subMetaPath, &meta)
		projectMap := make(map[string]projectIdent)
		for _, p := range meta.Projects {
			projectMap[p.Name] = p
		}

		// 1. SELECT TARGETS
		var targets []projectIdent

		if len(args) > 0 {
			// Select by name args
			for _, name := range args {
				if p, ok := projectMap[name]; ok {
					targets = append(targets, p)
				} else {
					// Fallback if local folder exists but not in meta
					if fileExists(filepath.Join(wd, name)) {
						targets = append(targets, projectIdent{Name: name})
					}
				}
			}
		} else if submitAll {
			targets = meta.Projects
		} else {
			// Interactive UI
			options := []huh.Option[string]{}
			for _, p := range meta.Projects {
				options = append(options, huh.NewOption(p.Name, p.Name))
			}
			var selected []string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("Select assignments to submit:").
						Options(options...).
						Value(&selected),
				),
			)
			if err := form.Run(); err != nil {
				return nil
			} // Cancelled
			for _, s := range selected {
				targets = append(targets, projectMap[s])
			}
		}

		if len(targets) == 0 {
			fmt.Printf("%s[INFO] No projects selected.%s\n", Yellow, Reset)
			return nil
		}

		// 2. MESSAGE INPUT
		finalMsg := submitMsg
		if finalMsg == "" {
			huh.NewInput().Title("Commit Message").Value(&finalMsg).Run()
		}
		if finalMsg == "" {
			finalMsg = "Update assignment"
		}

		// 3. EXECUTE
		var results []TaskResult
		var mu sync.Mutex

		title := fmt.Sprintf("Submitting %d project(s)...", len(targets))
		err = RunSpinner(title, func() error {
			var wg sync.WaitGroup
			for _, p := range targets {
				wg.Add(1)
				go func(proj projectIdent) {
					defer wg.Done()
					res := submitOneRepo(wd, proj.Name, finalMsg)
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

func submitOneRepo(wd, name, msg string) TaskResult {
	dir := filepath.Join(wd, name)
	if !fileExists(dir) {
		return TaskResult{Name: name, Status: "ERR", Message: "Folder missing"}
	}
	if !fileExists(filepath.Join(dir, ".git")) {
		return TaskResult{Name: name, Status: "ERR", Message: "Not a git repo"}
	}

	// 1. Add
	exec.Command("git", "-C", dir, "add", ".").Run()

	// 2. Check status
	out, _ := exec.Command("git", "-C", dir, "status", "--porcelain").Output()
	if len(out) > 0 {
		// Commit
		if err := exec.Command("git", "-C", dir, "commit", "-m", msg).Run(); err != nil {
			return TaskResult{Name: name, Status: "ERR", Message: "Commit failed"}
		}
	}

	// 3. Push
	if err := exec.Command("git", "-C", dir, "push", "--quiet").Run(); err != nil {
		return TaskResult{Name: name, Status: "ERR", Message: "Push failed"}
	}

	return TaskResult{Name: name, Status: "OK", Message: "Submitted"}
}

func init() {
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().BoolVar(&submitAll, "all", false, "Submit all")
	submitCmd.Flags().StringVarP(&submitMsg, "message", "m", "", "Commit message")
}
