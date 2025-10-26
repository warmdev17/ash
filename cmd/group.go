package cmd

import (
	"fmt"

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
			fmt.Println("Please use ash group fetch instead")
		case showGroups:
			fmt.Println("Please use ash group list instead")
		default:
			fmt.Println("Use ash group fetch/list")
			return nil
		}
		return nil
	},
}

// fetchAndSaveGroups executes `glab api "groups?owned=true&top_level_only=true" --paginate`,
// parses the JSON response, and writes ~/.config/ash/config.json.

func init() {
	rootCmd.AddCommand(groupCmd)
	groupCmd.Flags().BoolVarP(&listGroups, "get", "g", false, "Fetch and save all owned top-level GitLab groups")
	groupCmd.Flags().BoolVarP(&showGroups, "show", "s", false, "Show groups saved in ~/.config/ash/config.json")
}
