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

type glProject struct {
	ID            int64  `json:"id"`
	Path          string `json:"path"`
	Name          string `json:"name"`
	SSHURLToRepo  string `json:"ssh_url_to_repo"`
	HTTPURLToRepo string `json:"http_url_to_repo"`
}

type glGroup struct {
	ID                  int64  `json:"id"`
	Name                string `json:"name"`
	Path                string `json:"path"`
	MarkedForDeletionOn string `json:"marked_for_deletion_on"`
}

// ---------- Metadata types ----------

// Identifiers used in metadata files
type groupIdent struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
	Name string `json:"name"`
}

type projectIdent struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
	Name string `json:"name"`
}

type subgroupIdent struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
	Name string `json:"name"`
}

// Root group meta: .ash/group.json
type rootGroupMeta struct {
	Group     groupIdent      `json:"group"`
	Subgroups []subgroupIdent `json:"subgroup"`
}

// Subgroup meta: .ash/subgroup.json
type subgroupMeta struct {
	Group    groupIdent     `json:"group"`
	Projects []projectIdent `json:"projects"`
}
