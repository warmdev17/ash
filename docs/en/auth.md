# Auth Command

The `auth` command is used to authenticate with your GitLab instance.

## Login

Authenticate with a GitLab instance using a Personal Access Token (PAT).

### Usage

```bash
ash auth login [flags]
```

### Flags

- `-t, --token string`: Personal Access Token (required)
- `--hostname string`: GitLab hostname (default "git.rikkei.edu.vn")
- `--api-host string`: API host (host:port) (default "git.rikkei.edu.vn:443")
- `--api-protocol string`: API protocol (http|https) (default "https")
- `-g, --git-protocol string`: Git protocol (ssh|https) (default "https")

### Examples

Login with default settings (HTTPS):
```bash
ash auth login -t <your-token>
```

Login using SSH for Git operations:
```bash
ash auth login -t <your-token> -g ssh
```
