package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var proto string

var groupCloneCmd = &cobra.Command{
	Use:   "clone [name]",
	Short: "Clone an existing group's full hierarchy (subgroups + projects)",
	Example: `  ash group clone "CNTT2 - Spring 2025"
  ash group clone "My Organization" --git-proto ssh`,
	Args:          cobra.ExactArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		groupName := args[0]

		cfg, cfgPath, err := loadConfig()
		if err != nil {
			return err
		}
		grp, ok := findGroupByName(cfg, groupName)
		if !ok {
			return fmt.Errorf("group %q not found in %s; run 'ash group -g' first", groupName, cfgPath)
		}

		if proto != "ssh" && proto != "https" {
			proto = "https"
		}
		fmt.Printf("Cloning full hierarchy into %s (protocol: %s)\n", groupName, proto)

		if err := scaffoldLocalGroup(groupName, grp); err != nil {
			return err
		}
		return cloneGroupHierarchy(groupIdent{ID: grp.ID, Path: grp.Path}, groupName, proto, true)
	},
}

func init() {
	groupCmd.AddCommand(groupCloneCmd)
	groupCloneCmd.Flags().StringVar(&proto, "git-proto", "https", "Clone protocol (ssh|https)")
}
