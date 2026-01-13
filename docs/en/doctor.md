# Doctor Command

The `doctor` command checks your system for necessary dependencies and configuration.

## Usage

```bash
ash doctor [flags]
```

## Checks

The command verifies:
1. **OS**: Checks if the operating system is supported.
2. **Git**: Checks if `git` is installed.
3. **Glab**: Checks if `glab` CLI is installed and authenticated.
4. **Fzf**: Checks if `fzf` is installed (optional, but recommended).
5. **Config**: Checks if `ash` configuration file exists.

## Examples

```bash
ash doctor
```
