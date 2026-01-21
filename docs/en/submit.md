# Submit Command

The `submit` command automates the homework submission process. It updates your changes to GitLab.

## Usage

```bash
ash submit [flags]
```

## Description

This command performs the following actions:

1. Scans local directory for valid project folders (targets).
2. Prompts for commit message (if not provided via `-m`).
3. Executes `git add .`, `git commit`, and `git push` for each target.

## Flags

- `--all`: Submit all assignments in the current session (subgroup) non-interactively.
- `-m, --message string`: Commit message.

## Examples

Interactive submission:

```bash
ash submit
```

Submit with custom message:

```bash
ash submit -m "Complete Assignment 1"
```

Submit all assignments in the current directory:

```bash
ash submit --all
```
