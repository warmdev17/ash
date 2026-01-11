package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var projectListCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects in the current subgroup",
	RunE: func(cmd *cobra.Command, args []string) error {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		subMetaPath := filepath.Join(wd, ".ash", "subgroup.json")

		if !fileExists(subMetaPath) {
			return fmt.Errorf("not in a subgroup folder (.ash/subgroup.json missing)")
		}

		var meta subgroupMeta
		if err := readJSON(subMetaPath, &meta); err != nil {
			return err
		}

		if len(meta.Projects) == 0 {
			fmt.Println("No projects found in metadata.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tPATH")
		for _, p := range meta.Projects {
			fmt.Fprintf(w, "%d\t%s\t%s\n", p.ID, p.Name, p.Path)
		}
		w.Flush()
		return nil
	},
}

func init() {
	projectCmd.AddCommand(projectListCmd)
}
