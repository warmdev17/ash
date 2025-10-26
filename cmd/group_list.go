package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var groupListCmd = &cobra.Command{
	Use:           "list",
	Short:         "Display the saved groups",
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		return showSavedGroups()
	},
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
	groupCmd.AddCommand(groupListCmd)
}
