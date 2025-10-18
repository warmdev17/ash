package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

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
  ash submit -r 3,5,7 -c "Fix Baitap#"`,
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate flags
		if strings.TrimSpace(submitMsg) == "" {
			return errors.New("missing commit message: use -m or -c to provide message")
		}
		if submitAll && strings.TrimSpace(submitRange) != "" {
			return errors.New("use either --all or -r, not both")
		}
		if !submitAll && strings.TrimSpace(submitRange) == "" {
			return errors.New("must specify either --all or -r <list>")
		}

		// Must be inside a subgroup directory (has .ash/subgroup.json)
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		ashDir := filepath.Join(wd, ".ash")
		subMetaPath := filepath.Join(ashDir, "subgroup.json")
		if !fileExists(subMetaPath) {
			return errors.New("not inside a subgroup: .ash/subgroup.json not found")
		}

		// Read subgroup metadata for project list
		var meta subgroupMeta
		if err := readJSON(subMetaPath, &meta); err != nil {
			return fmt.Errorf("failed to read subgroup.json: %w", err)
		}
		if meta.Group.ID == 0 {
			return errors.New("invalid subgroup.json: missing group.id")
		}
		projects := meta.Projects

		// Build list of target repos
		selected := make(map[string]struct{})
		if submitAll {
			for _, p := range projects {
				selected[p.Name] = struct{}{}
			}
		} else {
			nums := parseNumList(submitRange)
			if len(nums) == 0 {
				return errors.New("no valid numbers found in -r (example: -r 1,3,5)")
			}
			for _, p := range projects {
				if n := trailingDigits(p.Name); n != "" {
					if _, ok := nums[n]; ok {
						selected[p.Name] = struct{}{}
					}
				}
			}
			if len(selected) == 0 {
				return fmt.Errorf("no repositories matched numbers: %q", submitRange)
			}
		}

		// Iterate through selected repos
		for _, p := range projects {
			if _, ok := selected[p.Name]; !ok {
				continue
			}
			repoDir := filepath.Join(wd, p.Name)
			if _, err := os.Stat(repoDir); err != nil {
				fmt.Printf("⏭️  Skip (not found): %s\n", p.Name)
				continue
			}
			if _, err := os.Stat(filepath.Join(repoDir, ".git")); err != nil {
				fmt.Printf("⏭️  Skip (not a git repo): %s\n", p.Name)
				continue
			}

			// Check for uncommitted changes
			st := exec.Command("git", "-C", repoDir, "status", "--porcelain")
			out, _ := st.Output()
			if len(out) == 0 {
				fmt.Printf("✔  Clean (no changes): %s\n", p.Name)
				continue
			}

			// Stage all changes
			if err := exec.Command("git", "-C", repoDir, "add", "-A").Run(); err != nil {
				fmt.Printf("❌ Add failed: %s (%v)\n", p.Name, err)
				continue
			}

			// Commit message: replace '#' with numeric suffix if available
			msg := submitMsg
			num := trailingDigits(p.Name)
			if strings.Contains(msg, "#") {
				msg = strings.ReplaceAll(msg, "#", num)
			}

			commit := exec.Command("git", "-C", repoDir, "commit", "-m", msg)
			if err := commit.Run(); err != nil {
				fmt.Printf("❌ Commit failed: %s (%v)\n", p.Name, err)
				continue
			}

			// Push to remote
			push := exec.Command("git", "-C", repoDir, "push", "--quiet")
			if err := push.Run(); err != nil {
				fmt.Printf("❌ Push failed: %s (%v)\n", p.Name, err)
				continue
			}
			fmt.Printf("🚀 Submitted: %s\n", p.Name)
		}

		fmt.Println("✅ All done.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
	submitCmd.Flags().BoolVar(&submitAll, "all", false, "Submit all repositories with changes")
	submitCmd.Flags().StringVarP(&submitRange, "repos", "r", "", "Submit only specific repo numbers (comma separated)")
	submitCmd.Flags().StringVarP(&submitMsg, "message", "m", "", "Commit message template (use '$' to insert repo number)")
	submitCmd.Flags().StringVarP(&submitMsg, "commit", "c", "", "Alias for --message")
}

// trailingDigits returns trailing numbers in a string (e.g. "Baitap12" -> "12").
func trailingDigits(s string) string {
	rx := regexp.MustCompile(`(\d+)$`)
	m := rx.FindStringSubmatch(strings.TrimSpace(s))
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

// parseNumList("3,5,7") -> {"3":{}, "5":{}, "7":{}}
func parseNumList(s string) map[string]struct{} {
	res := map[string]struct{}{}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if matched, _ := regexp.MatchString(`^\d+$`, part); matched {
			res[part] = struct{}{}
		}
	}
	return res
}
