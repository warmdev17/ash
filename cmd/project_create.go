package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	projectName   string
	projectCount  int
	projectPrefix string
	projectProto  string
)

var projectCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create multiple projects with prefix and count",
	Long: `Create one or more new projects in the current subgroup on GitLab.

Use the --count flag to specify how many projects to create,
and --prefix to define the project name prefix.

Examples:
  ash project create --count 5 --prefix Baitap
    → Creates projects: Baitap1, Baitap2, Baitap3, Baitap4, Baitap5

All created projects will also be recorded in subgroup.json for sync tracking.`,
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		if projectProto != "ssh" && projectProto != "https" {
			projectProto = "https"
		}

		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd failed: %w", err)
		}
		ashDir := filepath.Join(wd, ".ash")
		subMetaPath := filepath.Join(ashDir, "subgroup.json")
		if !fileExists(subMetaPath) {
			return errors.New("not in a subgroup folder: .ash/subgroup.json not found")
		}

		// read subgroup meta to get current subgroupo (namespace)
		var meta subgroupMeta
		if err := readJSON(subMetaPath, &meta); err != nil {
			return fmt.Errorf("parse subgroup.json failed: %w", err)
		}
		if meta.Group.ID == 0 {
			return errors.New("invalid subgroup.json: missing subgroup id")
		}

		// build names to create
		var names []string
		if strings.TrimSpace(projectName) != "" {
			if projectCount > 0 || strings.TrimSpace(projectPrefix) != "" {
				return errors.New("use either -n <Name> or -c <N> -p <prefix> ")
			}
			names = []string{projectName}
		} else {
			if projectCount <= 0 || strings.TrimSpace(projectPrefix) == "" {
				return errors.New("missing project count, please use -c <N>")
			}
			for i := 1; i <= projectCount; i++ {
				names = append(names, fmt.Sprintf("%s%d", projectPrefix, i))
			}
		}

		// create each project via glab API, then clone
		createdAny := false
		for _, name := range names {
			display := strings.TrimSpace(name)
			if display == "" {
				continue
			}
			path := slugify(display) // reuse helper from init.go

			fmt.Printf("%s Creating project: name=%q path=%q namespace_id=%d\n", icRun, display, path, meta.Group.ID)
			create := exec.Command("glab", "api", "-X", "POST", "/projects",
				"-f", "name="+display,
				"-f", "path="+path,
				"-f", "namespace_id="+strconv.FormatInt(meta.Group.ID, 10),
				"-f", "visibility=public",
			)
			out, err := create.Output()
			if err != nil {
				fmt.Printf("%s create failed for %s: %v\n", icErr, display, err)
				continue
			}

			// parse response using glProject (already defined in codebase)
			var pr glProject
			if err := json.Unmarshal(out, &pr); err != nil {
				fmt.Printf("%s parse create response failed for %s: %v\n", icErr, display, err)
				continue
			}
			if pr.ID == 0 {
				fmt.Printf("%s unexpected response for %s (no ID)\n", icErr, display)
				continue
			}
			fmt.Printf("%s created: id=%d name=%q path=%q\n", icOk, pr.ID, pr.Name, pr.Path)

			// clone locally into folder named by display Name
			dest := filepath.Join(wd, display)
			url := pr.SSHURLToRepo
			if repoProto == "https" {
				url = pr.HTTPURLToRepo
			}
			fmt.Printf("%s cloning → %s\n", icDownload, dest)
			clone := exec.Command("git", "clone", "--quiet", url, dest)
			if err := clone.Run(); err != nil {
				fmt.Printf("%s clone failed: %s (%v)\n", icErr, display, err)
				continue
			}
			fmt.Printf("%s cloned %s\n", icOk, display)
			createdAny = true
		}

		// refresh subgroup.json (take live snapshot) if at least one created
		if createdAny {
			fmt.Println("🗂  refreshing .ash/subgroup.json ...")
			projects, err := apiListProjects(meta.Group.ID) // lists projects for subgroup
			if err != nil {
				return err
			}
			prj := make([]projectIdent, 0, len(projects))
			for _, p := range projects {
				prj = append(prj, projectIdent{ID: p.ID, Path: p.Path, Name: p.Name})
			}
			meta.Projects = prj
			if err := writeSubgroupJSON(ashDir, meta); err != nil {
				return fmt.Errorf("write subgroup.json failed: %w", err)
			}
			fmt.Printf("%s Updated .ash/subgroup.json", icInfo)
		}

		fmt.Printf("%s Done.", icOk)
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectCreateCmd)

	projectCreateCmd.Flags().StringVarP(&projectName, "name", "n", "", "Name of project")
	projectCreateCmd.Flags().IntVarP(&projectCount, "count", "c", 0, "Count of project")
	projectCreateCmd.Flags().StringVarP(&projectPrefix, "prefix", "p", "", "Create multiple project with Prefix")
	projectCreateCmd.Flags().StringVarP(&projectProto, "proto", "g", "https", "Clone project ( https | ssh )")
}
