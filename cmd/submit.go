package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var (
	submitAll   bool
	submitRange string
	submitMsg   string
)

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Commit and push assignments in the current subgroup",
	Long: `Run inside a subgroup directory (one that has .ash/subgroup.json).

Examples:
  ash submit --all -m "Submit Session01 Baitap#"
  ash submit -r 3,5,7 -c "Fix Baitap#"
  ash submit (Interactive Mode)`,
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		// 1. Validate Flags
		if strings.TrimSpace(submitMsg) == "" {
			return errors.New("missing commit message: use -m or -c to provide message")
		}
		if submitAll && strings.TrimSpace(submitRange) != "" {
			return errors.New("use either --all or -r, not both")
		}

		// 2. Check Environment (Inside Subgroup?)
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		ashDir := filepath.Join(wd, ".ash")
		subMetaPath := filepath.Join(ashDir, "subgroup.json")
		if !fileExists(subMetaPath) {
			return errors.New("not inside a subgroup: .ash/subgroup.json not found")
		}

		// 3. Read Metadata
		var meta subgroupMeta
		if err := readJSON(subMetaPath, &meta); err != nil {
			return fmt.Errorf("failed to read subgroup.json: %w", err)
		}
		projects := meta.Projects

		// 4. Select Projects
		selectedProjects := []projectIdent{}

		if submitAll {
			selectedProjects = projects
		} else if len(submitRange) > 0 {
			nums := parseNumList(submitRange)
			if len(nums) == 0 {
				return errors.New("no valid numbers found in -r (example: -r 1,3,5)")
			}
			for _, p := range projects {
				if n := trailingDigits(p.Name); n != "" {
					if _, ok := nums[n]; ok {
						selectedProjects = append(selectedProjects, p)
					}
				}
			}
			if len(selectedProjects) == 0 {
				return fmt.Errorf("no repositories matched numbers: %q", submitRange)
			}
		} else {
			// Interactive Mode
			options := []huh.Option[string]{}
			projectMap := make(map[string]projectIdent)

			for _, p := range projects {
				options = append(options, huh.NewOption(p.Name, p.Name))
				projectMap[p.Name] = p
			}

			var selectedNames []string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewMultiSelect[string]().
						Title("Select assignments to submit (Space to select, Enter to confirm):").
						Options(options...).
						Value(&selectedNames),
				),
			)

			err := form.Run()
			if err != nil {
				return err
			}
			if len(selectedNames) == 0 {
				fmt.Println("No projects selected. Aborting.")
				return nil
			}

			for _, name := range selectedNames {
				selectedProjects = append(selectedProjects, projectMap[name])
			}
		}

		// 5. Concurrent Submission
		fmt.Printf("\nProcessing %d repositories...\n", len(selectedProjects))

		var wg sync.WaitGroup
		results := make(chan string, len(selectedProjects))
		sem := make(chan struct{}, 5) // Max 5 concurrent processes

		for _, p := range selectedProjects {
			wg.Add(1)
			go func(proj projectIdent) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				resMsg := processRepo(wd, proj, submitMsg)
				results <- resMsg
			}(p)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		for msg := range results {
			fmt.Print(msg)
		}

		fmt.Println("All done.")
		return nil
	},
}

func processRepo(wd string, p projectIdent, msgTemplate string) string {
	repoDir := filepath.Join(wd, p.Name)

	if _, err := os.Stat(repoDir); err != nil {
		return fmt.Sprintf("[%s] Skip: not found\n", p.Name)
	}
	if _, err := os.Stat(filepath.Join(repoDir, ".git")); err != nil {
		return fmt.Sprintf("[%s] Skip: not a git repo\n", p.Name)
	}

	st := exec.Command("git", "-C", repoDir, "status", "--porcelain")
	out, _ := st.Output()
	if len(out) == 0 {
		return fmt.Sprintf("[%s] Clean: no changes\n", p.Name)
	}

	if err := exec.Command("git", "-C", repoDir, "add", "-A").Run(); err != nil {
		return fmt.Sprintf("[%s] Error: Add failed (%v)\n", p.Name, err)
	}

	msg := msgTemplate
	num := trailingDigits(p.Name)
	if strings.Contains(msg, "#") {
		msg = strings.ReplaceAll(msg, "#", num)
	}

	commit := exec.Command("git", "-C", repoDir, "commit", "-m", msg)
	if err := commit.Run(); err != nil {
		return fmt.Sprintf("[%s] Error: Commit failed (%v)\n", p.Name, err)
	}

	push := exec.Command("git", "-C", repoDir, "push", "--quiet")
	if err := push.Run(); err != nil {
		return fmt.Sprintf("[%s] Error: Push failed (%v)\n", p.Name, err)
	}

	return fmt.Sprintf("[%s] Submitted successfully\n", p.Name)
}

func init() {
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().BoolVar(&submitAll, "all", false, "Submit all repositories with changes")
	submitCmd.Flags().StringVarP(&submitRange, "repos", "r", "", "Submit only specific repo numbers (comma separated)")
	submitCmd.Flags().StringVarP(&submitMsg, "message", "m", "", "Commit message template (use '#' to insert repo number)")
	submitCmd.Flags().StringVarP(&submitMsg, "commit", "c", "", "Alias for --message")
}

func trailingDigits(s string) string {
	rx := regexp.MustCompile(`(\d+)$`)
	m := rx.FindStringSubmatch(strings.TrimSpace(s))
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

func parseNumList(s string) map[string]struct{} {
	res := map[string]struct{}{}
	rx := regexp.MustCompile(`^\d+$`)

	for part := range strings.SplitSeq(s, ",") {
		part = strings.TrimSpace(part)
		if rx.MatchString(part) {
			res[part] = struct{}{}
		}
	}
	return res
}
