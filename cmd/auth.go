// Copyright © 2025 warmdev warmdevofficial@gmail.com
package cmd

import (
	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate and manage login sesssions",
}

func init() {
	rootCmd.AddCommand(authCmd)
}
