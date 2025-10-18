package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// FLAGS
var (
	initName     string // -n : create a new top-level group on GitLab
	initDir      string // -d : target directory; defaults to --name/--existing
	existingName string // -e : use an existing group and clone its hierarchy
	initGitProto string // --git-proto : ssh|https for cloning in existing mode (default ssh)
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a GitLab group or clone an existing group's full hierarchy",
	Long: `Two modes:
1) Create mode:   ash init -n "<New Group Name>" [-d <dir>]
   - Create a top-level GitLab group (visibility=public) via glab API
   - Immediately re-sync ~/.config/ash/config.json (same as 'ash group -g')
   - Scaffold <dir> (or --name if -d omitted) and write <dir>/.ash/group.json with empty "subgroup": []

2) Existing mode: ash init -e "<Existing Group Name>" [-d <dir>] [--git-proto ssh|https]
   - Use group info from ~/.config/ash/config.json
   - Scaffold <dir> (or --existing if -d omitted)
   - At root: write .ash/group.json with "subgroup": [{id,path,name},...]
   - For every subgroup: write .ash/subgroup.json with "projects": [{id,path,name},...]
   - Clone repos into folders named by their GitLab Name (capitalized)
   - Cloning is quiet by default (only prints success/errors per repo)`,
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		hasNew := strings.TrimSpace(initName) != ""
		hasExisting := strings.TrimSpace(existingName) != ""
		if hasNew == hasExisting {
			return errors.New("choose exactly one mode: -n <name> to create OR -e <name> to clone existing")
		}

		cfg, cfgPath, err := loadConfig()
		if err != nil {
			return err
		}

		if hasExisting {
			// -------- EXISTING MODE --------
			grp, ok := findGroupByName(cfg, existingName)
			if !ok {
				return fmt.Errorf("group %q not found in %s; run 'ash group -g' first", existingName, cfgPath)
			}
			target := pickDir(existingName, initDir)

			// scaffold folder + empty group.json at root
			if err := scaffoldLocalGroup(target, grp); err != nil {
				return err
			}
			if initGitProto != "ssh" && initGitProto != "https" {
				initGitProto = "https"
			}
			fmt.Printf("Cloning full hierarchy into %s (protocol: %s)\n", target, initGitProto)
			// recursive clone & meta writing — root=true
			return cloneGroupHierarchy(groupIdent{ID: grp.ID, Path: grp.Path}, target, initGitProto, true)
		}

		// -------- CREATE MODE --------
		// If already listed in config, reuse
		if g, ok := findGroupByName(cfg, initName); ok {
			fmt.Printf("Group already exists: name=%q id=%d path=%q\n", g.Name, g.ID, g.Path)
			return scaffoldLocalGroup(pickDir(initName, initDir), g) // writes .ash/group.json with empty subgroup
		}

		// Create new group via glab
		slug := slugify(initName)
		fmt.Printf("Creating group via glab: name=%q path=%q visibility=public\n", initName, slug)

		glabCmd := exec.Command("glab", "api", "-X", "POST", "/groups",
			"-f", "name="+initName,
			"-f", "path="+slug,
			"-f", "visibility=public",
		)
		out, err := glabCmd.Output()
		if err != nil {
			return fmt.Errorf("glab failed: %w", err)
		}

		var created struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
			Path string `json:"path"`
		}
		if err := json.Unmarshal(out, &created); err != nil {
			return fmt.Errorf("failed to parse glab response: %w", err)
		}
		if created.ID == 0 || created.Path == "" {
			return fmt.Errorf("unexpected glab response: missing id/path")
		}

		fmt.Printf("Created group: id=%d path=%q\n", created.ID, created.Path)

		// Re-sync config.json (same as `ash group -g`)
		if err := fetchAndSaveGroups(); err != nil {
			return fmt.Errorf("resync config after create failed: %w", err)
		}

		// Reload config to find canonical entry
		cfg2, cfgPath2, err := loadConfig()
		if err != nil {
			return fmt.Errorf("reload config failed: %w", err)
		}
		g, ok := findGroupByName(cfg2, initName)
		if !ok {
			g = GitLabGroup{ID: created.ID, Name: initName, Path: created.Path}
			fmt.Printf("Warning: created group not found in %s after resync; using API response\n", cfgPath2)
		} else {
			fmt.Printf("Config synced: name=%q id=%d path=%q (%s)\n", g.Name, g.ID, g.Path, cfgPath2)
		}

		// scaffold root with empty subgroup — writes .ash/group.json
		return scaffoldLocalGroup(pickDir(initName, initDir), g)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&initName, "name", "n", "", "Create a new top-level group with this display name")
	initCmd.Flags().StringVarP(&existingName, "existing", "e", "", "Use an existing group and clone its hierarchy")
	initCmd.Flags().StringVarP(&initDir, "dir", "d", "", "Target directory (defaults to --name or --existing)")
	initCmd.Flags().StringVar(&initGitProto, "git-proto", "https", "Clone protocol for existing mode (ssh|https)")
}

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

// ---------- Metadata types ----------

// Identifiers used in metadata files
type groupIdent struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
}

type projectIdent struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
	Name string `json:"name"`
}

type subgroupIdent struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
	Name string `json:"name"`
}

// Root group meta: .ash/group.json
type rootGroupMeta struct {
	Group     groupIdent      `json:"group"`
	Subgroups []subgroupIdent `json:"subgroup"`
}

// Subgroup meta: .ash/subgroup.json
type subgroupMeta struct {
	Group    groupIdent     `json:"group"`
	Projects []projectIdent `json:"projects"`
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

// ---------------- cloning logic (existing mode) ----------------

type glProject struct {
	ID            int64  `json:"id"`
	Path          string `json:"path"`
	Name          string `json:"name"`
	SSHURLToRepo  string `json:"ssh_url_to_repo"`
	HTTPURLToRepo string `json:"http_url_to_repo"`
}

type glGroup struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
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
		fmt.Printf("Found %d project(s) in group %d\n", len(projects), group.ID)
	}
	for _, p := range projects {
		url := p.SSHURLToRepo
		if proto == "https" {
			url = p.HTTPURLToRepo
		}
		dest := filepath.Join(rootDir, p.Name) // use Name for directory

		if _, err := os.Stat(dest); err == nil {
			fmt.Printf("Skip (exists): %s\n", dest)
			continue
		}
		clone := exec.Command("git", "clone", "--quiet", url, dest)
		if err := clone.Run(); err != nil {
			fmt.Printf("clone failed: %s (%v)\n", p.Name, err)
			continue
		}
		fmt.Printf("cloned %s\n", p.Name)
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
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("glab api projects failed: %w", err)
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
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("glab api subgroups failed: %w", err)
	}
	var sgs []glGroup
	if err := json.Unmarshal(out, &sgs); err != nil {
		return nil, fmt.Errorf("parse subgroups failed: %w", err)
	}
	return sgs, nil
}
