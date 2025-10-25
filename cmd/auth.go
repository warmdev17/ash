// Copyright © 2025 warmdev warmdevofficial@gmail.com
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate and manage login sesssions",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("auth called")
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
