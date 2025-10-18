package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var (
	listGroups bool // -g / --get : fetch from GitLab via glab and save
	showGroups bool // -s / --show: display saved groups from config
)

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage GitLab groups",
	Long: `Fetch your owned top-level GitLab groups using glab CLI and store them in ~/.config/ash/config.json,
or display the saved groups.

Examples:
  ash group -g   # Fetch and save groups
  ash group -s   # Show groups saved in ~/.config/ash/config.json`,
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		switch {
		case listGroups:
			return fetchAndSaveGroups()
		case showGroups:
			return showSavedGroups()
		default:
			fmt.Println("Use -g to fetch groups or -s to show saved groups.")
			return nil
		}
	},
}

// fetchAndSaveGroups executes `glab api "groups?owned=true&top_level_only=true" --paginate`,
// parses the JSON response, and writes ~/.config/ash/config.json.
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

	fmt.Printf(" Saved %d groups to %s\n", len(groups), configPath)
	return nil
}

// showSavedGroups reads ~/.config/ash/config.json and prints a table of groups (ID, NAME, PATH).
func showSavedGroups() error {
	// Resolve config path
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("failed to get user config dir: %w", err)
	}
	configPath := filepath.Join(configDir, "ash", "config.json")

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("config file not found, run 'ash group -g' first")
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var cfg AshConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Empty case
	if len(cfg.Groups) == 0 {
		fmt.Println("No groups saved in config.")
		return nil
	}

	// Pretty print using tabwriter (explicitly ignore write/flush errors to satisfy linters)
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tNAME\tPATH")
	for _, g := range cfg.Groups {
		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\n", g.ID, g.Name, g.Path)
	}
	_ = w.Flush()
	return nil
}

func init() {
	rootCmd.AddCommand(groupCmd)
	groupCmd.Flags().BoolVarP(&listGroups, "get", "g", false, "Fetch and save all owned top-level GitLab groups")
	groupCmd.Flags().BoolVarP(&showGroups, "show", "s", false, "Show groups saved in ~/.config/ash/config.json")
}
