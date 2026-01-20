package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var subgroupCloneCmd = &cobra.Command{
	Use:   "clone [name]",
	Short: "Clone a specific subgroup into the current group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		wd, _ := os.Getwd()
		groupMetaPath := filepath.Join(wd, ".ash", "group.json")

		if !fileExists(groupMetaPath) {
			return fmt.Errorf("not in a group root (.ash/group.json missing)")
		}

		var meta rootGroupMeta
		if err := readJSON(groupMetaPath, &meta); err != nil {
			return err
		}

		// Find Remote Subgroup by Name (need to search API inside parent)
		// Or search in current meta?
		// Usually if we want to clone, maybe we manually deleted it or it wasn't there?
		// Let's search API.
		found, sg, err := findSubgroupByPath(meta.Group.ID, slugify(name))
		if err != nil {
			return err
		}
		if !found {
			// Try fuzzy match or exact name match
			all, _ := apiListSubgroups(meta.Group.ID)
			for _, s := range all {
				if s.Name == name {
					sg = s
					found = true
					break
				}
			}
		}

		if !found {
			return fmt.Errorf("subgroup %q not found on GitLab under group %s", name, meta.Group.Name)
		}

		fmt.Printf("Cloning subgroup: %s (ID: %d)\n", sg.Name, sg.ID)

		// Create Folder
		targetDir := filepath.Join(wd, sg.Name)
		os.MkdirAll(targetDir, 0o755)

		// Determine Protocol
		proto := "https"
		cfg, _, _ := loadConfig()
		if cfg.GitProtocol != "" {
			proto = cfg.GitProtocol
		}

		// Recurse Clone
		err = RunSpinner(fmt.Sprintf("Cloning subgroup %s", sg.Name), func() error {
			if err := cloneGroupHierarchy(groupIdent{ID: sg.ID, Path: sg.Path, Name: sg.Name}, targetDir, proto, false); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}

		// Update Parent Meta if missing
		exists := false
		for _, s := range meta.Subgroups {
			if s.ID == sg.ID {
				exists = true
				break
			}
		}
		if !exists {
			meta.Subgroups = append(meta.Subgroups, subgroupIdent{ID: sg.ID, Path: sg.Path, Name: sg.Name})
			writeGroupJSON(filepath.Join(wd, ".ash"), meta)
		}

		fmt.Println("[OK] Subgroup cloned.")
		return nil
	},
}

func init() {
	subgroupCmd.AddCommand(subgroupCloneCmd)
}
