package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	subgroupCreateDir  string // --dir: local folder name (optional, default = subgroup display name)
	subgroupVisibility string // --visibility: public|internal|private (default public)
)

var subgroupCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a subgroup under the current group and scaffold a local folder",
	Example: `  cd MyGroup
  ash subgroup create "Session 1"
  ash subgroup create "Session 1" --dir S1
  ash subgroup create "Secret Lab" --visibility private`,
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// 1) Must be inside a group root (has .ash/group.json, NOT .ash/subgroup.json)
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd failed: %w", err)
		}
		ashDir := filepath.Join(wd, ".ash")
		groupMetaPath := filepath.Join(ashDir, "group.json")
		subMetaPath := filepath.Join(ashDir, "subgroup.json")

		if !fileExists(groupMetaPath) {
			return fmt.Errorf("not in a group root: %s not found", groupMetaPath)
		}
		if fileExists(subMetaPath) {
			return fmt.Errorf("this looks like a subgroup folder (found %s); run from the group root", subMetaPath)
		}

		// 2) Read current group meta (need parent group ID)
		var meta rootGroupMeta
		if err := readJSON(groupMetaPath, &meta); err != nil {
			return fmt.Errorf("parse group.json failed: %w", err)
		}
		if meta.Group.ID == 0 {
			return fmt.Errorf("invalid group.json: missing group.id")
		}

		// 3) Prepare subgroup slug/path
		path := slugify(name)

		// 3a) Preflight: check if subgroup already exists under the parent (avoid 409)
		existed, existedSG, err := findSubgroupByPath(meta.Group.ID, path)
		if err != nil {
			return fmt.Errorf("check existing subgroup failed: %w", err)
		}
		if existed {
			fmt.Printf("Subgroup already exists on GitLab: id=%d name=%q path=%q\n", existedSG.ID, existedSG.Name, existedSG.Path)
			return scaffoldAndLinkSubgroup(wd, &meta, existedSG.ID, existedSG.Name, existedSG.Path)
		}

		// 4) Create subgroup on GitLab via glab (default visibility = public)
		if subgroupVisibility == "" {
			subgroupVisibility = "public"
		}

		var created struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
			Path string `json:"path"`
		}

		err = RunSpinner(fmt.Sprintf("Creating subgroup %s", name), func() error {
			argsPost := []string{
				"api", "-X", "POST", "/groups",
				"-f", "name=" + name,
				"-f", "path=" + path,
				"-f", fmt.Sprintf("parent_id=%d", meta.Group.ID),
				"-f", "visibility=" + subgroupVisibility, // default public
			}
			glabCmd := exec.Command("glab", argsPost...)
			out, err := glabCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("glab create subgroup failed: %v\n%s", err, string(out))
			}

			if err := json.Unmarshal(out, &created); err != nil {
				return fmt.Errorf("failed to parse glab response: %w", err)
			}
			if created.ID == 0 || created.Path == "" || created.Name == "" {
				return fmt.Errorf("unexpected glab response: missing id/name/path")
			}
			return nil
		})
		if err != nil {
			return err
		}

		fmt.Printf("Created subgroup: id=%d name=%q path=%q\n", created.ID, created.Name, created.Path)

		// 5) Local scaffold + update parent group.json
		return scaffoldAndLinkSubgroup(wd, &meta, created.ID, created.Name, created.Path)
	},
}

func init() {
	subgroupCmd.AddCommand(subgroupCreateCmd)
	subgroupCreateCmd.Flags().StringVar(&subgroupCreateDir, "dir", "", "Custom local directory name (optional)")
	subgroupCreateCmd.Flags().StringVar(&subgroupVisibility, "visibility", "public", "Subgroup visibility: public|internal|private (default: public)")
}

// ---------- helpers (local to subgroup create) ----------

func scaffoldAndLinkSubgroup(wd string, meta *rootGroupMeta, sgID int64, sgName, sgPath string) error {
	ashDir := filepath.Join(wd, ".ash")

	// Scaffold folder: <dirOrName>/.ash/subgroup.json (empty projects)
	dirName := strings.TrimSpace(subgroupCreateDir)
	if dirName == "" {
		dirName = sgName
	}
	subDir := filepath.Join(wd, dirName)
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		return fmt.Errorf("create subgroup dir %q: %w", subDir, err)
	}

	sgMeta := subgroupMeta{
		Group:    groupIdent{ID: sgID, Path: sgPath},
		Projects: []projectIdent{},
	}
	if err := writeSubgroupJSON(filepath.Join(subDir, ".ash"), sgMeta); err != nil {
		return fmt.Errorf("write subgroup.json failed: %w", err)
	}
	fmt.Printf("Scaffolded: %s\n", subDir)
	fmt.Printf("Wrote: %s\n", filepath.Join(subDir, ".ash", "subgroup.json"))

	// Update parent's .ash/group.json (dedupe by Name, case-insensitive)
	lower := strings.ToLower(sgName)
	exists := false
	for _, s := range meta.Subgroups {
		if strings.ToLower(s.Name) == lower {
			exists = true
			break
		}
	}
	if !exists {
		meta.Subgroups = append(meta.Subgroups, subgroupIdent{
			ID:   sgID,
			Path: sgPath,
			Name: sgName,
		})
		if err := writeGroupJSON(ashDir, *meta); err != nil {
			return fmt.Errorf("update group.json failed: %w", err)
		}
		fmt.Println("Updated parent .ash/group.json")
	} else {
		fmt.Println("Subgroup already listed in parent group.json; no change")
	}

	fmt.Println("Done.")
	return nil
}
