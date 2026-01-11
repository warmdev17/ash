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

// --- COLORS (ANSI) ---
var (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
)

// --- TASK RESULTS ---

type TaskResult struct {
	Name    string // Object Name (e.g. Exercise1)
	Status  string // OK, ERR, NEW, SKIP
	Message string // Details (Cloned, Push failed...)
}

func PrintResults(results []TaskResult) {
	if len(results) == 0 {
		return
	}
	fmt.Println() // Spacer
	for _, res := range results {
		var color, tag string
		switch res.Status {
		case "OK":
			color = Green
			tag = "[OK]"
		case "ERR":
			color = Red
			tag = "[ERR]"
		case "NEW":
			color = Cyan
			tag = "[NEW]"
		case "SKIP":
			color = Yellow
			tag = "[SKIP]"
		default:
			color = Gray
			tag = "[INFO]"
		}
		// Format: [TAG] Name (căn lề 20) : Message
		fmt.Printf("%s%-7s %-20s : %s%s\n", color, tag, res.Name, res.Message, Reset)
	}
	fmt.Println()
}

// --- FILE / JSON HELPERS ---

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

func writeJSON(path string, v any) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, _ := json.MarshalIndent(v, "", "  ")
	return os.WriteFile(path, b, 0o644)
}

// Specific wrapper for writing group/subgroup json to match legacy code
func writeGroupJSON(ashDir string, meta rootGroupMeta) error {
	return writeJSON(filepath.Join(ashDir, "group.json"), meta)
}

func writeSubgroupJSON(ashDir string, meta subgroupMeta) error {
	return writeJSON(filepath.Join(ashDir, "subgroup.json"), meta)
}

// findSubgroupByPath returns (true, glGroup, nil) if a subgroup with the given slug/path exists under parentID.
func findSubgroupByPath(parentID int64, slug string) (bool, glGroup, error) {
	sgs, err := apiListSubgroups(parentID)
	if err != nil {
		return false, glGroup{}, err
	}
	for _, sg := range sgs {
		// match by Path (slug); Name can vary in case/spacing
		if strings.EqualFold(sg.Path, slug) {
			return true, sg, nil
		}
	}
	return false, glGroup{}, nil
}

func loadConfig() (AshConfig, string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return AshConfig{}, "", err
	}
	ashDir := filepath.Join(configDir, "ash")
	cfgPath := filepath.Join(ashDir, "config.json")

	var cfg AshConfig
	if !fileExists(cfgPath) {
		return AshConfig{}, cfgPath, nil
	}
	err = readJSON(cfgPath, &cfg)
	return cfg, cfgPath, err
}

func saveConfig(path string, cfg AshConfig) error {
	return writeJSON(path, cfg)
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "-")
	re := regexp.MustCompile(`[^a-z0-9-]`)
	s = re.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// --- LOGIC HELPERS (Group/Clone) ---

func findGroupByName(cfg AshConfig, name string) (GitLabGroup, bool) {
	want := strings.ToLower(strings.TrimSpace(name))
	for _, g := range cfg.Groups {
		if strings.ToLower(g.Name) == want {
			return g, true
		}
	}
	return GitLabGroup{}, false
}

// scaffoldLocalGroup creates directory and basic .ash/group.json
func scaffoldLocalGroup(dir string, g GitLabGroup) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}
	ashMeta := filepath.Join(dir, ".ash")
	if err := os.MkdirAll(ashMeta, 0o755); err != nil {
		return fmt.Errorf("failed to create .ash directory: %w", err)
	}
	meta := rootGroupMeta{
		Group:     groupIdent{ID: g.ID, Path: g.Path, Name: g.Name},
		Subgroups: []subgroupIdent{},
	}
	if err := writeGroupJSON(ashMeta, meta); err != nil {
		return err
	}
	fmt.Printf("%s[OK] Scaffolded local group: %s%s\n", Green, dir, Reset)
	return nil
}

// --- API / CLONE HELPERS ---

func apiListProjects(groupID int64) ([]glProject, error) {
	url := fmt.Sprintf("groups/%d/projects?per_page=100&simple=true", groupID)
	out, err := exec.Command("glab", "api", url, "--paginate").Output()
	if err != nil {
		return nil, err
	}
	var prjs []glProject
	if err := json.Unmarshal(out, &prjs); err != nil {
		return nil, err
	}
	return prjs, nil
}

func apiListSubgroups(groupID int64) ([]glGroup, error) {
	url := fmt.Sprintf("groups/%d/subgroups?per_page=100", groupID)
	out, err := exec.Command("glab", "api", url, "--paginate").Output()
	if err != nil {
		return nil, err
	}
	var sgs []glGroup
	if err := json.Unmarshal(out, &sgs); err != nil {
		return nil, err
	}
	return sgs, nil
}

// cloneGroupHierarchy recursively clones the entire group/subgroup hierarchy
func cloneGroupHierarchy(group groupIdent, rootDir, proto string, isRoot bool) error {
	// 1. Fetch data
	subgroups, err := apiListSubgroups(group.ID)
	if err != nil {
		return err
	}
	projects, err := apiListProjects(group.ID)
	if err != nil {
		return err
	}

	// 2. Write metadata
	ashDir := filepath.Join(rootDir, ".ash")
	if isRoot {
		sgIdents := make([]subgroupIdent, 0, len(subgroups))
		for _, sg := range subgroups {
			sgIdents = append(sgIdents, subgroupIdent{ID: sg.ID, Path: sg.Path, Name: sg.Name})
		}
		writeGroupJSON(ashDir, rootGroupMeta{Group: group, Subgroups: sgIdents})
	} else {
		prjIdents := make([]projectIdent, 0, len(projects))
		for _, p := range projects {
			prjIdents = append(prjIdents, projectIdent{ID: p.ID, Path: p.Path, Name: p.Name})
		}
		writeSubgroupJSON(ashDir, subgroupMeta{Group: group, Projects: prjIdents})
	}

	// 3. Clone Projects (Batch)
	if len(projects) > 0 {
		fmt.Printf("Syncing %d projects in %s...\n", len(projects), rootDir)
		for _, p := range projects {
			url := p.HTTPURLToRepo
			if proto == "ssh" {
				url = p.SSHURLToRepo
			}
			dest := filepath.Join(rootDir, p.Name)
			if !fileExists(dest) {
				exec.Command("git", "clone", "--quiet", url, dest).Run()
			}
		}
	}

	// 4. Recurse Subgroups
	for _, sg := range subgroups {
		sgDir := filepath.Join(rootDir, sg.Name)
		os.MkdirAll(sgDir, 0o755)
		cloneGroupHierarchy(groupIdent{ID: sg.ID, Path: sg.Path}, sgDir, proto, false)
	}

	return nil
}
