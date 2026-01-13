# Submit Command

The `submit` command automates the homework submission process. It pushes your changes to GitLab and creates a Merge Request.

## Usage

```bash
ash submit [flags]
```

## Description

This command performs the following actions:
1. Adds all changes (`git add .`)
2. Commits changes with a message (`git commit -m "Submit homework"`)
3. Pushes to the current branch (`git push origin <branch>`)
4. Creates a Merge Request (MR) targeting the default branch.

## Flags

- `-m, --message string`: Custom commit message (default "Submit homework")
- `-t, --title string`: Title of the Merge Request (default is the last commit message)
- `-d, --description string`: Description of the Merge Request
- `--draft`: Create the MR as a Draft
- `-l, --label strings`: Add labels to the MR
- `-a, --assignee strings`: Assign users to the MR
- `-r, --reviewer strings`: Request reviewers for the MR

## Examples

Basic submission:
```bash
ash submit
```

Submission with custom message and labels:
```bash
ash submit -m "Complete Assignment 1" -l "homework,backend"
```
