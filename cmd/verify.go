package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var (
	Token    string
	Host     string
	APIHost  string
	APIProto string
	GitProto string
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Authenticate with GitLab using a Personal Access Token (PAT)",
	Long: `Authenticate to a GitLab instance using your personal access token.

Examples:
  ash verify -t <token> -g ssh
  ash verify -t <token> -g https`,
	SilenceUsage:  false, // Print usage when missing flags
	SilenceErrors: true,  // Cobra won't print errors automatically

	RunE: func(cmd *cobra.Command, args []string) error {
		// --- Validate token ---
		if Token == "" {
			return errors.New("missing --token, example: ash verify -t <PAT>")
		}

		// --- Validate git protocol ---
		if GitProto != "ssh" && GitProto != "https" {
			return fmt.Errorf("invalid git protocol: %s (allowed: ssh or https)", GitProto)
		}

		// --- Build glab login command ---
		fmt.Printf("Running command: glab [auth login --hostname %s --token ***redacted*** --api-host %s --api-protocol %s --git-protocol %s]\n",
			Host, APIHost, APIProto, GitProto)

		login := exec.Command("glab",
			"auth", "login",
			"--hostname", Host,
			"--token", Token,
			"--api-host", APIHost,
			"--api-protocol", APIProto,
			"--git-protocol", GitProto,
		)
		login.Stdout = os.Stdout
		login.Stderr = os.Stderr

		if err := login.Run(); err != nil {
			return fmt.Errorf("%s glab login failed: %w", icErr, err)
		}

		// --- Verify authentication status ---
		status := exec.Command("glab", "auth", "status", "--hostname", Host, "--show-token")
		status.Stdout = os.Stdout
		status.Stderr = os.Stderr

		// --- Set default host to git.rikkei.edu.vn ---
		defaultHost := exec.Command("glab", "config", "set", "-g", "host", "git.rikkei.edu.vn")
		defaultHost.Stdout = os.Stdout
		defaultHost.Stderr = os.Stderr

		if err := status.Run(); err != nil {
			return fmt.Errorf("auth status check failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	verifyCmd.Flags().StringVarP(&Token, "token", "t", "", "Personal Access Token (PAT) for authentication")
	// verifyCmd.Flags().Lookup("token").NoOptDefVal = "" // allow `-t` with no argument

	verifyCmd.Flags().StringVar(&Host, "hostname", "git.rikkei.edu.vn", "GitLab hostname")
	verifyCmd.Flags().StringVar(&APIHost, "api-host", "git.rikkei.edu.vn:443", "API host (host:port)")
	verifyCmd.Flags().StringVar(&APIProto, "api-protocol", "https", "API protocol (http|https)")
	verifyCmd.Flags().StringVarP(&GitProto, "git-protocol", "g", "https", "Git protocol (ssh|https)")
}
