package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system requirements and status",
	Long:  `Check if git, glab are installed and verify authentication status.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running system health check...")
		fmt.Println("------------------------------")

		hasErrors := false

		// 1. Check Git
		if path, err := exec.LookPath("git"); err == nil {
			fmt.Printf("[PASS] Git found at: %s\n", path)
		} else {
			fmt.Println("[FAIL] Git is NOT installed.")
			hasErrors = true
		}

		// 2. Check Glab CLI
		if path, err := exec.LookPath("glab"); err == nil {
			fmt.Printf("[PASS] Glab CLI found at: %s\n", path)

			// 3. Check Authentication (Only if glab is installed)
			// Check if logged in to any host
			out, err := exec.Command("glab", "auth", "status").CombinedOutput()
			if err == nil {
				fmt.Println("[PASS] Glab authentication: Logged in.")
			} else {
				fmt.Printf("[WARN] Glab authentication issue:\n%s\n", string(out))
				fmt.Println("       Run 'ash auth login' to fix.")
			}
		} else {
			fmt.Println("[FAIL] Glab CLI is NOT installed (Required).")
			hasErrors = true
		}

		fmt.Println("------------------------------")
		if hasErrors {
			fmt.Println("Please install missing dependencies to use ash.")
		} else {
			fmt.Println("System is ready.")
		}
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
