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
	syncProto string // --proto ssh|https
	syncDry   bool   // --dry-run
	syncClean bool   // --clean
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize the current group/subgroup folder with GitLab",
	Long: `Run inside a group or subgroup directory.
- If .ash/group.json exists (root group): fetch subgroups from GitLab, update group.json,
  create missing subgroup folders (by Name), and write their .ash/subgroup.json.
- If .ash/subgroup.json exists: fetch projects from GitLab, update subgroup.json,
  clone new repos (folder named by project Name), and optionally remove missing repos with --clean.

Flags:
  --proto ssh|https   Clone protocol (default: https)
  --dry-run           Preview actions (no writes/clone/delete)
  --clean             Remove local folders that no longer exist on GitLab`,
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		if syncProto != "ssh" && syncProto != "https" {
			syncProto = "https"
		}

		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("getwd failed: %w", err)
		}
		ashDir := filepath.Join(wd, ".ash")

		groupMetaPath := filepath.Join(ashDir, "group.json")
		subMetaPath := filepath.Join(ashDir, "subgroup.json")

		groupExists := fileExists(groupMetaPath)
		subExists := fileExists(subMetaPath)

		if !groupExists && !subExists {
			return errors.New("no .ash/group.json or .ash/subgroup.json found; run inside a group or subgroup folder")
		}
		if groupExists && subExists {
			return errors.New("both .ash/group.json and .ash/subgroup.json exist; please keep only one (root vs subgroup)")
		}

		if groupExists {
			return syncRootGroup(wd, groupMetaPath)
		}
		return syncSubgroup(wd, subMetaPath)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().StringVar(&syncProto, "proto", "https", "Clone protocol (ssh|https)")
	syncCmd.Flags().BoolVar(&syncDry, "dry-run", false, "Preview changes only; do not write/clone/delete")
	syncCmd.Flags().BoolVar(&syncClean, "clean", false, "Remove local folders not present on GitLab")
}

// ---------------- Root group sync (.ash/group.json) ----------------

func syncRootGroup(rootDir, metaPath string) error {
	fmt.Println("🔍 Reading .ash/group.json ...")
	var meta rootGroupMeta
	if err := readJSON(metaPath, &meta); err != nil {
		return fmt.Errorf("parse group.json failed: %w", err)
	}
	if meta.Group.ID == 0 {
		return errors.New("invalid group.json: missing group.id")
	}

	// Fetch latest subgroups from GitLab
	fmt.Printf("🌐 Fetching subgroups for group %d ...\n", meta.Group.ID)
	sgs, err := apiListSubgroups(meta.Group.ID)
	if err != nil {
		return err
	}

	// Build new subgroup list
	newSubgroups := make([]subgroupIdent, 0, len(sgs))
	for _, sg := range sgs {
		newSubgroups = append(newSubgroups, subgroupIdent{
			ID:   sg.ID,
			Path: sg.Path,
			Name: sg.Name,
		})
	}

	// Compute diffs against existing meta
	oldMap := map[string]subgroupIdent{}
	for _, x := range meta.Subgroups {
		oldMap[strings.ToLower(x.Name)] = x
	}
	newMap := map[string]subgroupIdent{}
	for _, x := range newSubgroups {
		newMap[strings.ToLower(x.Name)] = x
	}

	// Added subgroups
	for name, sg := range newMap {
		if _, ok := oldMap[name]; !ok {
			if syncDry {
				fmt.Printf("➕ [dry-run] would add subgroup folder: %s\n", name)
			} else {
				fmt.Printf("➕ Add subgroup folder: %s\n", name)
				sgDir := filepath.Join(rootDir, name)
				if err := os.MkdirAll(sgDir, 0o755); err != nil {
					return fmt.Errorf("create subgroup dir %q: %w", sgDir, err)
				}
				// Also write its .ash/subgroup.json (projects snapshot)
				if err := writeSubgroupSnapshot(sgDir, groupIdent{ID: sg.ID, Path: sg.Path}); err != nil {
					return err
				}
			}
		}
	}

	// Removed subgroups
	for name := range oldMap {
		if _, ok := newMap[name]; !ok {
			sgDir := filepath.Join(rootDir, name)
			if _, err := os.Stat(sgDir); os.IsNotExist(err) {
				continue // thư mục không tồn tại sẵn
			}
			if syncClean {
				if syncDry {
					fmt.Printf("🗑️  [dry-run] would remove missing subgroup: %s\n", sgDir)
				} else {
					fmt.Printf("🗑️  Remove missing subgroup: %s\n", sgDir)
					if err := os.RemoveAll(sgDir); err != nil {
						return fmt.Errorf("remove subgroup dir %q: %w", sgDir, err)
					}
				}
			} else {
				fmt.Printf("⚠️  Missing on GitLab: %s (use --clean to remove)\n", sgDir)
			}
		}
	}

	// Update group.json with fresh subgroup list
	if syncDry {
		fmt.Println("✨ [dry-run] would update .ash/group.json with latest subgroups")
		return nil
	}
	meta.Subgroups = newSubgroups
	if err := writeGroupJSON(filepath.Join(rootDir, ".ash"), meta); err != nil {
		return fmt.Errorf("write group.json failed: %w", err)
	}
	fmt.Println("✨ Updated .ash/group.json successfully")
	return nil
}

// Write .ash/subgroup.json for a subgroup directory using live projects snapshot.
// This also clones any missing repos if they don't exist locally? No—this is root sync.
// We only take a snapshot here (no clone) to keep root-sync fast/minimal.
func writeSubgroupSnapshot(subgroupDir string, gid groupIdent) error {
	ashDir := filepath.Join(subgroupDir, ".ash")
	if err := os.MkdirAll(ashDir, 0o755); err != nil {
		return fmt.Errorf("create .ash for subgroup: %w", err)
	}
	projects, err := apiListProjects(gid.ID)
	if err != nil {
		return err
	}
	prj := make([]projectIdent, 0, len(projects))
	for _, p := range projects {
		prj = append(prj, projectIdent{ID: p.ID, Path: p.Path, Name: p.Name})
	}
	meta := subgroupMeta{
		Group:    gid,
		Projects: prj,
	}
	return writeSubgroupJSON(ashDir, meta)
}

// ---------------- Subgroup sync (.ash/subgroup.json) ----------------

func syncSubgroup(rootDir, metaPath string) error {
	fmt.Println("🔍 Reading .ash/subgroup.json ...")
	var meta subgroupMeta
	if err := readJSON(metaPath, &meta); err != nil {
		return fmt.Errorf("parse subgroup.json failed: %w", err)
	}
	if meta.Group.ID == 0 {
		return errors.New("invalid subgroup.json: missing group.id")
	}

	// Fetch latest projects from GitLab
	fmt.Printf("🌐 Fetching projects for group %d ...\n", meta.Group.ID)
	projects, err := apiListProjects(meta.Group.ID)
	if err != nil {
		return err
	}

	// Build new project list
	newProjects := make([]projectIdent, 0, len(projects))
	for _, p := range projects {
		newProjects = append(newProjects, projectIdent{
			ID:   p.ID,
			Path: p.Path,
			Name: p.Name,
		})
	}

	// Diff vs existing meta
	oldMap := map[string]projectIdent{} // by Name (directory is Name)
	for _, x := range meta.Projects {
		oldMap[x.Name] = x
	}
	newMap := map[string]projectIdent{}
	for _, x := range newProjects {
		newMap[x.Name] = x
	}

	// Added projects → clone
	for name := range newMap {
		if _, ok := oldMap[name]; !ok {
			dest := filepath.Join(rootDir, name)
			if syncDry {
				fmt.Printf("➕ [dry-run] would clone: %s\n", name)
				continue
			}
			fmt.Printf("➕ Clone: %s\n", name)

			// find full project info to get clone URL
			var full glProject
			for _, p := range projects {
				if p.Name == name {
					full = p
					break
				}
			}
			url := full.SSHURLToRepo
			if syncProto == "https" {
				url = full.HTTPURLToRepo
			}

			clone := exec.Command("git", "clone", "--quiet", url, dest)
			if err := clone.Run(); err != nil {
				fmt.Printf("❌ clone failed: %s (%v)\n", name, err)
				continue
			}
			fmt.Printf("✅ cloned %s\n", name)
		}
	}

	// Removed projects → delete folder (with --clean), else warn
	for name := range oldMap {
		if _, ok := newMap[name]; !ok {
			dest := filepath.Join(rootDir, name)
			if syncClean {
				if syncDry {
					fmt.Printf("🗑️  [dry-run] would remove: %s\n", dest)
				} else {
					fmt.Printf("🗑️  Remove: %s\n", dest)
					if err := os.RemoveAll(dest); err != nil {
						return fmt.Errorf("remove repo dir %q: %w", dest, err)
					}
				}
			} else {
				fmt.Printf("⚠️  Missing on GitLab: %s (use --clean to remove)\n", dest)
			}
		}
	}

	// Update subgroup.json with fresh projects
	if syncDry {
		fmt.Println("✨ [dry-run] would update .ash/subgroup.json with latest projects")
		return nil
	}
	meta.Projects = newProjects
	if err := writeSubgroupJSON(filepath.Join(rootDir, ".ash"), meta); err != nil {
		return fmt.Errorf("write subgroup.json failed: %w", err)
	}
	fmt.Println("✨ Updated .ash/subgroup.json successfully")
	return nil
}

// ---------------- utilities ----------------

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
