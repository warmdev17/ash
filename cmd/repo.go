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
	repoName   string // -n :
	repoCount  int    // -c :
	repoPrefix string // -p :
	repoProto  string // --proto ssh|https
)

var repoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Create repository(ies) inside the current subgroup and clone locally",
	Long: `Run inside a subgroup directory (one that has .ash/subgroup.json).

Examples:
  ash repo -n Baitap1
  ash repo -c 10 -p Baitap
  ash repo -c 5 -p Lab --proto ssh`,
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		// validate proto
		if repoProto != "ssh" && repoProto != "https" {
			repoProto = "https"
		}

		// must be inside a subgroup (has .ash/subgroup.json)
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd failed: %w", err)
		}
		ashDir := filepath.Join(wd, ".ash")
		subMetaPath := filepath.Join(ashDir, "subgroup.json")
		if !fileExists(subMetaPath) { // helper exists in sync.go
			return errors.New("not in a subgroup folder: .ash/subgroup.json not found")
		}

		// read subgroup meta to get current subgroup (namespace) ID
		var meta subgroupMeta
		if err := readJSON(subMetaPath, &meta); err != nil { // readJSON helper
			return fmt.Errorf("parse subgroup.json failed: %w", err)
		}
		if meta.Group.ID == 0 {
			return errors.New("invalid subgroup.json: missing group.id")
		}

		// build names to create
		var names []string
		if strings.TrimSpace(repoName) != "" {
			if repoCount > 0 || strings.TrimSpace(repoPrefix) != "" {
				return errors.New("use either -n <Name> OR -c <N> -p <prefix>")
			}
			names = []string{repoName}
		} else {
			if repoCount <= 0 || strings.TrimSpace(repoPrefix) == "" {
				return errors.New("missing -n <Name> or (-c <N> -p <prefix>)")
			}
			for i := 1; i <= repoCount; i++ {
				names = append(names, fmt.Sprintf("%s%d", repoPrefix, i))
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
	rootCmd.AddCommand(repoCmd)
	repoCmd.Flags().StringVarP(&repoName, "name", "n", "", "Create a single repo with this display name")
	repoCmd.Flags().IntVarP(&repoCount, "count", "c", 0, "Create N repos with a numeric suffix")
	repoCmd.Flags().StringVarP(&repoPrefix, "prefix", "p", "", "Prefix for batch repo creation")
	repoCmd.Flags().StringVar(&repoProto, "proto", "https", "Clone protocol (ssh|https)")
}
