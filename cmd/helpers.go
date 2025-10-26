package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// ---------------- helpers ----------------

func pickDir(name, dirFlag string) string {
	if strings.TrimSpace(dirFlag) != "" {
		return dirFlag
	}
	return name
}

// slugify converts a display name to a lowercased, URL-safe path (a-z0-9-)
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	re := regexp.MustCompile(`[^a-z0-9-]`)
	s = re.ReplaceAllString(s, "-")
	re2 := regexp.MustCompile(`-+`)
	s = re2.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "group"
	}
	return s
}

func loadConfig() (AshConfig, string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return AshConfig{}, "", fmt.Errorf("failed to get user config dir: %w", err)
	}
	ashDir := filepath.Join(configDir, "ash")
	cfgPath := filepath.Join(ashDir, "config.json")

	var cfg AshConfig
	if err := os.MkdirAll(ashDir, 0o755); err != nil {
		return AshConfig{}, "", fmt.Errorf("failed to create config dir: %w", err)
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := saveConfig(cfgPath, cfg); err != nil {
				return AshConfig{}, "", err
			}
			return AshConfig{}, cfgPath, nil
		}
		return AshConfig{}, "", fmt.Errorf("failed to read config file: %w", err)
	}
	if len(data) > 0 {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return AshConfig{}, "", fmt.Errorf("failed to parse config file: %w", err)
		}
	}
	return cfg, cfgPath, nil
}

func saveConfig(path string, cfg AshConfig) error {
	b, _ := json.MarshalIndent(cfg, "", "  ")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return nil
}

func findGroupByName(cfg AshConfig, name string) (GitLabGroup, bool) {
	want := strings.ToLower(strings.TrimSpace(name))
	for _, g := range cfg.Groups {
		if strings.ToLower(g.Name) == want {
			return g, true
		}
	}
	return GitLabGroup{}, false
}

func writeGroupJSON(ashDir string, meta rootGroupMeta) error {
	if err := os.MkdirAll(ashDir, 0o755); err != nil {
		return fmt.Errorf("failed to create .ash: %w", err)
	}
	b, _ := json.MarshalIndent(meta, "", "  ")
	return os.WriteFile(filepath.Join(ashDir, "group.json"), b, 0o644)
}

func writeSubgroupJSON(ashDir string, meta subgroupMeta) error {
	if err := os.MkdirAll(ashDir, 0o755); err != nil {
		return fmt.Errorf("failed to create .ash: %w", err)
	}
	b, _ := json.MarshalIndent(meta, "", "  ")
	return os.WriteFile(filepath.Join(ashDir, "subgroup.json"), b, 0o644)
}

func apiListProjects(groupID int64) ([]glProject, error) {
	url := fmt.Sprintf("groups/%d/projects?per_page=100&simple=true", groupID)
	cmd := exec.Command("glab", "api", url, "--paginate")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("glab api projects failed: %v\n%s", err, string(out))
	}
	var prjs []glProject
	if err := json.Unmarshal(out, &prjs); err != nil {
		return nil, fmt.Errorf("parse projects failed: %w", err)
	}
	return prjs, nil
}

func apiListSubgroups(groupID int64) ([]glGroup, error) {
	url := fmt.Sprintf("groups/%d/subgroups?per_page=100", groupID)
	cmd := exec.Command("glab", "api", url, "--paginate")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("glab api subgroups failed: %v\n%s", err, string(out))
	}
	var sgs []glGroup
	if err := json.Unmarshal(out, &sgs); err != nil {
		return nil, fmt.Errorf("parse subgroups failed: %w", err)
	}
	return sgs, nil
}

// cloneGroupHierarchy recursively clones repos and subgroups.
// At root: writes .ash/group.json (group + subgroups).
// In subgroups: writes .ash/subgroup.json (group + projects).
// Uses "Name" as local folder name for better readability.
func cloneGroupHierarchy(group groupIdent, rootDir, proto string, isRoot bool) error {
	// 1) At this level, fetch projects (for subgroups) and subgroups (for root)
	projects, err := apiListProjects(group.ID)
	if err != nil {
		return err
	}
	subgroups, err := apiListSubgroups(group.ID)
	if err != nil {
		return err
	}

	// 1a) Write metadata file for this level
	if isRoot {
		// Root: write group.json with subgroups
		sgIdents := make([]subgroupIdent, 0, len(subgroups))
		for _, sg := range subgroups {
			sgIdents = append(sgIdents, subgroupIdent{
				ID:   sg.ID,
				Path: sg.Path,
				Name: sg.Name,
			})
		}
		if err := writeGroupJSON(filepath.Join(rootDir, ".ash"), rootGroupMeta{
			Group:     group,
			Subgroups: sgIdents,
		}); err != nil {
			return err
		}
	} else {
		// Subgroup: write subgroup.json with projects
		prjIdents := make([]projectIdent, 0, len(projects))
		for _, p := range projects {
			prjIdents = append(prjIdents, projectIdent{
				ID:   p.ID,
				Path: p.Path,
				Name: p.Name,
			})
		}
		if err := writeSubgroupJSON(filepath.Join(rootDir, ".ash"), subgroupMeta{
			Group:    group,
			Projects: prjIdents,
		}); err != nil {
			return err
		}
	}

	// 2) Clone projects at this level (both root and subgroups may have projects)
	if len(projects) > 0 {
		// Batch clone with spinner (no noisy per-repo logs)
		title := fmt.Sprintf("Cloning %d project(s) in %s …", len(projects), rootDir)
		var okList, badList []CloneResult

		_ = runWithSpinner(title, func() error {
			for _, p := range projects {
				url := p.SSHURLToRepo
				if proto == "https" {
					url = p.HTTPURLToRepo
				}
				dest := filepath.Join(rootDir, p.Name) // use Name for directory

				if _, err := os.Stat(dest); err == nil {
					// existed locally → treat as success but skip
					okList = append(okList, CloneResult{Name: p.Name, URL: url, Dest: dest})
					continue
				}
				res := cloneOneRepo(url, dest, p.Name)
				if res.Err != nil {
					badList = append(badList, res)
				} else {
					okList = append(okList, res)
				}
			}
			return nil
		})

		printCloneSummary(okList, badList)
	}

	// 3) Recurse into subgroups
	for _, sg := range subgroups {
		sgDir := filepath.Join(rootDir, sg.Name) // subgroup dir by display Name
		if err := os.MkdirAll(sgDir, 0o755); err != nil {
			return fmt.Errorf("failed to create subgroup dir %q: %w", sgDir, err)
		}
		fmt.Printf("Enter subgroup: %s\n", sg.Name)
		if err := cloneGroupHierarchy(groupIdent{ID: sg.ID, Path: sg.Path}, sgDir, proto, false /*subgroup*/); err != nil {
			fmt.Printf("error in subgroup %s: %v\n", sg.Name, err)
		}
	}
	return nil
}

// scaffoldLocalGroup creates <dir> and writes .ash/group.json (root groups use group.json with empty "subgroup": [])
func scaffoldLocalGroup(dir string, g GitLabGroup) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}
	ashMeta := filepath.Join(dir, ".ash")
	if err := os.MkdirAll(ashMeta, 0o755); err != nil {
		return fmt.Errorf("failed to create .ash directory: %w", err)
	}
	meta := rootGroupMeta{
		Group:     groupIdent{ID: g.ID, Path: g.Path},
		Subgroups: []subgroupIdent{},
	}
	if err := writeGroupJSON(ashMeta, meta); err != nil {
		return err
	}
	fmt.Printf("Scaffolded: %s\n", dir)
	fmt.Printf("Wrote: %s\n", filepath.Join(ashMeta, "group.json"))
	return nil
}
