package cmd

import (
	"fmt"
	"os/exec"
	"strconv"

	"github.com/spf13/cobra"
)

var groupDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a top-level GitLab group",
	Long: `Delete a top-level GitLab group.
This command will:
1. Delete the group from GitLab (using glab API).
2. Remove the group from your local configuration file.

It does NOT delete the local directory on your disk.`,
	Example:       `  ash group delete "IT103_K25_LeTrungHieu"`,
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		groupName := args[0]
		return deleteGroup(groupName)
	},
}

func init() {
	groupCmd.AddCommand(groupDeleteCmd)
}

func deleteGroup(name string) error {
	cfg, cfgPath, err := loadConfig()
	if err != nil {
		return err
	}

	g, ok := findGroupByName(cfg, name)
	if !ok {
		return fmt.Errorf("group %q not found in config (%s)", name, cfgPath)
	}

	fmt.Printf("Deleting group %q (ID: %d) from GitLab...\n", g.Name, g.ID)

	// glab api -X DELETE /groups/:id
	glabCmd := exec.Command("glab", "api", "-X", "DELETE", "/groups/"+strconv.FormatInt(g.ID, 10))
	if out, err := glabCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to delete group on GitLab: %s (%w)", string(out), err)
	}

	fmt.Printf("%s[OK] Group deleted from GitLab.%s\n", Green, Reset)

	// Remove from config
	newGroups := []GitLabGroup{}
	for _, grp := range cfg.Groups {
		if grp.ID != g.ID {
			newGroups = append(newGroups, grp)
		}
	}
	cfg.Groups = newGroups

	if err := saveConfig(cfgPath, cfg); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Printf("Updated config: %s\n", cfgPath)
	return nil
}
