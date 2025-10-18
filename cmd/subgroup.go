package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	subgroupName string
	subgroupDir  string
)

var subgroupCmd = &cobra.Command{
	Use:   "subgroup",
	Short: "Create a subgroup under the current group and scaffold local folder",
	Long: `Run inside a group root directory (one that has .ash/group.json).
Example:
  cd IT108_K25_LeTrungHieu
  ash subgroup -n "Session1" [-d "S1"]  # -d để đặt tên thư mục local tuỳ ý

It will:
  • Create the subgroup on GitLab (visibility=public) under the current group.
  • Create local folder ./<dir or Name>.
  • Write ./<dir or Name>/.ash/subgroup.json (empty projects).
  • Update parent's .ash/group.json (append new subgroup).`,
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		name := strings.TrimSpace(subgroupName)
		if name == "" {
			return errors.New("missing -n <subgroup name>")
		}

		// 1) Must be inside a group root (has .ash/group.json, and NOT .ash/subgroup.json)
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd failed: %w", err)
		}
		ashDir := filepath.Join(wd, ".ash")
		groupMetaPath := filepath.Join(ashDir, "group.json")
		subMetaPath := filepath.Join(ashDir, "subgroup.json")

		if !fileExists(groupMetaPath) {
			return errors.New("not in a group root: .ash/group.json not found")
		}
		if fileExists(subMetaPath) {
			return errors.New("this looks like a subgroup folder (found .ash/subgroup.json); run from the group root")
		}

		// 2) Read current group meta (need parent group ID)
		var meta rootGroupMeta
		if err := readJSON(groupMetaPath, &meta); err != nil {
			return fmt.Errorf("parse group.json failed: %w", err)
		}
		if meta.Group.ID == 0 {
			return errors.New("invalid group.json: missing group.id")
		}

		// 3) Create subgroup on GitLab via glab
		path := slugify(name) // path/slug an toàn
		fmt.Printf("Creating subgroup via glab: name=%q path=%q parent_id=%d visibility=public\n", name, path, meta.Group.ID)

		glabCmd := exec.Command("glab", "api", "-X", "POST", "/groups",
			"-f", "name="+name,
			"-f", "path="+path,
			"-f", fmt.Sprintf("parent_id=%d", meta.Group.ID),
			"-f", "visibility=public",
		)
		out, err := glabCmd.Output()
		if err != nil {
			return fmt.Errorf("glab create subgroup failed: %w", err)
		}

		var created struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
			Path string `json:"path"`
		}
		if err := json.Unmarshal(out, &created); err != nil {
			return fmt.Errorf("failed to parse glab response: %w", err)
		}
		if created.ID == 0 || created.Path == "" || created.Name == "" {
			return fmt.Errorf("unexpected glab response: missing id/name/path")
		}
		fmt.Printf("Created subgroup: id=%d name=%q path=%q\n", created.ID, created.Name, created.Path)

		// 4) Local scaffold: <dirOrName>/.ash/subgroup.json (empty projects)
		dirName := strings.TrimSpace(subgroupDir)
		if dirName == "" {
			dirName = created.Name // mặc định dùng display Name
		}
		subDir := filepath.Join(wd, dirName)
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			return fmt.Errorf("create subgroup dir %q: %w", subDir, err)
		}

		sgMeta := subgroupMeta{
			Group:    groupIdent{ID: created.ID, Path: created.Path},
			Projects: []projectIdent{},
		}
		if err := writeSubgroupJSON(filepath.Join(subDir, ".ash"), sgMeta); err != nil {
			return fmt.Errorf("write subgroup.json failed: %w", err)
		}
		fmt.Printf("Scaffolded: %s\n", subDir)
		fmt.Printf("Wrote: %s\n", filepath.Join(subDir, ".ash", "subgroup.json"))

		// 5) Update parent's .ash/group.json (avoid duplicate by Name-insensitive)
		lower := strings.ToLower(created.Name)
		exists := false
		for _, s := range meta.Subgroups {
			if strings.ToLower(s.Name) == lower {
				exists = true
				break
			}
		}
		if !exists {
			meta.Subgroups = append(meta.Subgroups, subgroupIdent{
				ID:   created.ID,
				Path: created.Path,
				Name: created.Name,
			})
			if err := writeGroupJSON(ashDir, meta); err != nil {
				return fmt.Errorf("update group.json failed: %w", err)
			}
			fmt.Println("✨ Updated parent .ash/group.json")
		} else {
			fmt.Println("ℹ️ Subgroup already listed in parent group.json; no change")
		}

		fmt.Println("✅ Done.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(subgroupCmd)
	subgroupCmd.Flags().StringVarP(&subgroupName, "name", "n", "", "Subgroup display name to create under current group")
	subgroupCmd.Flags().StringVarP(&subgroupDir, "dir", "d", "", "Custom local directory name (optional)")
}
