package cmd

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var groupCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new top-level GitLab group",
	Example: `  ash group create "CNTT2 - Spring 2025"
  ash group create "My Organization"`,
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		groupName := args[0]
		return createNewGroup(groupName, groupName)
	},
}

func init() {
	groupCmd.AddCommand(groupCreateCmd)
}

func createNewGroup(name, dir string) error {
	cfg, _, err := loadConfig()
	if err != nil {
		return err
	}

	if g, ok := findGroupByName(cfg, name); ok {
		fmt.Printf("Group already exists: name=%q id=%d path=%q\n", g.Name, g.ID, g.Path)
		return scaffoldLocalGroup(dir, g)
	}

	slug := slugify(name)
	fmt.Printf("Creating group via glab: name=%q path=%q visibility=public\n", name, slug)

	glabCmd := exec.Command("glab", "api", "-X", "POST", "/groups",
		"-f", "name="+name,
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

	if err := fetchAndSaveGroups(); err != nil {
		return fmt.Errorf("resync config after create failed: %w", err)
	}

	cfg2, cfgPath2, err := loadConfig()
	if err != nil {
		return fmt.Errorf("reload config failed: %w", err)
	}

	g, ok := findGroupByName(cfg2, name)
	if !ok {
		g = GitLabGroup{ID: created.ID, Name: name, Path: created.Path}
		fmt.Printf("Warning: created group not found in %s after resync; using API response\n", cfgPath2)
	} else {
		fmt.Printf("Config synced: name=%q id=%d path=%q (%s)\n", g.Name, g.ID, g.Path, cfgPath2)
	}

	return scaffoldLocalGroup(dir, g)
}
