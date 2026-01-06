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

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with GitLab using a Personal Access Token (PAT)",
	Long: `Authenticate to a GitLab instance using your personal access token.

Examples:
  ash auth login -t <token> -g ssh (login with ssh git protocol)
  ash auth login -t <token> -g https (login with https git protocol)`,
	SilenceUsage:  false,
	SilenceErrors: true,

	RunE: func(cmd *cobra.Command, args []string) error {
		// -- Validate token --
		if Token == "" {
			return errors.New("missing --token, example: ash auth login -t <personal access token>")
		}

		// -- Validate Git Protocol --
		if GitProto != "ssh" && GitProto != "https" {
			return fmt.Errorf("invalid git protocol: %s ( allow ssh or https)", GitProto)
		}

		login := exec.Command("glab", "auth", "login",
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
		status := exec.Command("glab", "auth", "status", "--hostname", Host)
		status.Stdout = os.Stdout
		status.Stderr = os.Stderr

		// --- Set default host to git.rikkei.edu.vn ---
		defaultHost := exec.Command("glab", "config", "set", "-g", "host", Host)
		defaultHost.Stdout = os.Stdout
		defaultHost.Stderr = os.Stderr

		if err := status.Run(); err != nil {
			return fmt.Errorf("auth status check failed: %w", err)
		}

		if err := defaultHost.Run(); err != nil {
			return fmt.Errorf("failed to set default host: %w", err)
		}
		fmt.Printf("%s Default GitLab host set to %s\n", icOk, Host)

		return nil
	},
}

func init() {
	authCmd.AddCommand(loginCmd)

	loginCmd.Flags().StringVarP(&Token, "token", "t", "", "Token")
	loginCmd.Flags().StringVar(&Host, "hostname", "git.rikkei.edu.vn", "GitLab hostname")
	loginCmd.Flags().StringVar(&APIHost, "api-host", "git.rikkei.edu.vn:443", "API host (host:port)")
	loginCmd.Flags().StringVar(&APIProto, "api-protocol", "https", "API protocol (http|https)")
	loginCmd.Flags().StringVarP(&GitProto, "git-protocol", "g", "https", "Git protocol (ssh|https)")
}
