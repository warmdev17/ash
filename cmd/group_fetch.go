package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var groupFetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch your owned top-level GitLab groups using glab CLI",

	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		return fetchAndSaveGroups()
	},
}

func fetchAndSaveGroups() error {
	// Execute the glab command (requires prior `glab auth login`)
	glabCmd := exec.Command("glab", "api", "groups?owned=true&top_level_only=true", "--paginate")
	out, err := glabCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to execute glab: %w", err)
	}

	// Parse JSON response -> []GitLabGroup
	var groups []GitLabGroup
	if err := json.Unmarshal(out, &groups); err != nil {
		return fmt.Errorf("failed to parse glab output: %w", err)
	}

	// Handle empty result gracefully
	if len(groups) == 0 {
		fmt.Println("No owned top-level groups found.")
		return nil
	}

	// Resolve config path: ~/.config/ash/config.json
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config dir: %w", err)
	}
	ashDir := filepath.Join(configDir, "ash")
	configPath := filepath.Join(ashDir, "config.json")

	// Ensure directory exists
	if err := os.MkdirAll(ashDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	// Write pretty JSON to config.json
	buf, _ := json.MarshalIndent(AshConfig{Groups: groups}, "", "  ")
	if err := os.WriteFile(configPath, buf, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("%s Saved %d groups to %s\n", icOk, len(groups), configPath)
	return nil
}

func init() {
	groupCmd.AddCommand(groupFetchCmd)
}
