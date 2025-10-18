package cmd

// GitLabGroup represents a GitLab group (used across multiple subcommands)
type GitLabGroup struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// AshConfig defines ~/.config/ash/config.json structure
type AshConfig struct {
	Groups []GitLabGroup `json:"groups"`
}
