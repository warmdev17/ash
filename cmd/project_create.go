package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/cobra"
)

var (
	createProjectCount  int
	createProjectPrefix string
	createProjectProto  string
)

var projectCreateCmd = &cobra.Command{
	Use:   "create [names...]",
	Short: "Create projects (Interactive or Batch)",
	Long: `Create one or more projects in the current subgroup.

Examples:
  ash project create BaiTap1 BaiTap2
  ash project create -c 5 -p Lab`,
	SilenceUsage: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		// 0. Resolve Protocol Default from Config
		if !cmd.Flags().Changed("proto") {
			cfg, _, _ := loadConfig()
			if cfg.GitProtocol != "" {
				createProjectProto = cfg.GitProtocol
			}
		}

		// 1. Env Check
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		ashDir := filepath.Join(wd, ".ash")
		subMetaPath := filepath.Join(ashDir, "subgroup.json")

		if !fileExists(subMetaPath) {
			return fmt.Errorf("not in a subgroup folder (.ash/subgroup.json missing)")
		}

		var meta subgroupMeta
		if err := readJSON(subMetaPath, &meta); err != nil {
			return fmt.Errorf("read metadata failed: %w", err)
		}

		// 2. Determine List
		var names []string
		if len(args) > 0 {
			names = args
		} else {
			if createProjectCount <= 0 || createProjectPrefix == "" {
				return fmt.Errorf("missing args: usage 'ash project create Name' or '-c 5 -p Prefix'")
			}
			for i := 1; i <= createProjectCount; i++ {
				names = append(names, fmt.Sprintf("%s%d", createProjectPrefix, i))
			}
		}

		// 3. EXECUTE
		var results []TaskResult
		var mu sync.Mutex // Để append kết quả an toàn

		title := fmt.Sprintf("Creating %d project(s)...", len(names))

		err = RunSpinner(title, func() error {
			for _, rawName := range names {
				display := strings.TrimSpace(rawName)
				if display == "" {
					continue
				}

				res := createOneProject(wd, meta.Group.ID, display, createProjectProto)

				mu.Lock()
				results = append(results, res)
				mu.Unlock()
			}

			// Refresh Metadata Silent
			refreshProjectMeta(ashDir, meta.Group.ID)
			return nil
		})
		if err != nil {
			return err
		}
		PrintResults(results)
		return nil
	},
}

func createOneProject(wd string, groupID int64, name string, proto string) TaskResult {
	path := slugify(name)

	createCmd := exec.Command("glab", "api", "/projects", "-X", "POST",
		"-f", "name="+name,
		"-f", "path="+path,
		"-f", "namespace_id="+strconv.FormatInt(groupID, 10),
		"-f", "visibility=public",
	)

	out, err := createCmd.Output()
	if err != nil {
		return TaskResult{Name: name, Status: "ERR", Message: "GitLab create failed"}
	}

	var pr glProject
	if err := json.Unmarshal(out, &pr); err != nil {
		return TaskResult{Name: name, Status: "ERR", Message: "Parse response failed"}
	}

	// B. Clone
	dest := filepath.Join(wd, name)
	repoURL := pr.HTTPURLToRepo
	if proto == "ssh" {
		repoURL = pr.SSHURLToRepo
	}

	if err := exec.Command("git", "clone", "--quiet", repoURL, dest).Run(); err != nil {
		return TaskResult{Name: name, Status: "ERR", Message: "Created but Clone failed"}
	}

	return TaskResult{Name: name, Status: "OK", Message: "Ready"}
}

func refreshProjectMeta(ashDir string, groupID int64) {
	prjs, _ := apiListProjects(groupID)
	if len(prjs) > 0 {
		var newMeta subgroupMeta
		readJSON(filepath.Join(ashDir, "subgroup.json"), &newMeta)

		idents := make([]projectIdent, 0, len(prjs))
		for _, p := range prjs {
			idents = append(idents, projectIdent{ID: p.ID, Path: p.Path, Name: p.Name})
		}
		newMeta.Projects = idents
		writeJSON(filepath.Join(ashDir, "subgroup.json"), newMeta)
	}
}

func init() {
	projectCmd.AddCommand(projectCreateCmd)
	projectCreateCmd.Flags().IntVarP(&createProjectCount, "count", "c", 0, "Number of projects")
	projectCreateCmd.Flags().StringVarP(&createProjectPrefix, "prefix", "p", "", "Prefix for batch creation")
	projectCreateCmd.Flags().StringVarP(&createProjectProto, "proto", "g", "https", "Protocol (https/ssh)")
}
