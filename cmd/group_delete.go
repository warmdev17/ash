package cmd

import (
	"github.com/spf13/cobra"
)

var groupDeleteCmd = &cobra.Command{
	Use:     "delete [delete]",
	Short:   "Create a new top-level GitLab group",
	Example: `ash group delete "IT103_K25_LeTrungHieu"`,
}
