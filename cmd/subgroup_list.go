package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var subgroupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List subgroups in the current group",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		groupMetaPath := filepath.Join(wd, ".ash", "group.json")

		if !fileExists(groupMetaPath) {
			return fmt.Errorf("not in a group root (.ash/group.json missing)")
		}

		var meta rootGroupMeta
		if err := readJSON(groupMetaPath, &meta); err != nil {
			return err
		}

		if len(meta.Subgroups) == 0 {
			fmt.Println("No subgroups found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tPATH")
		for _, sg := range meta.Subgroups {
			fmt.Fprintf(w, "%d\t%s\t%s\n", sg.ID, sg.Name, sg.Path)
		}
		w.Flush()
		return nil
	},
}

func init() {
	subgroupCmd.AddCommand(subgroupListCmd)
}
