# Group Command

The `group` command manages GitLab groups.

## Usage

```bash
ash group [command]
```

## Available Commands

### list

List all groups available to you.

```bash
ash group list
```

### create

Create a new group in GitLab.

```bash
ash group create
```

**Flags:**
- `-n, --name string`: Name of the group
- `-p, --path string`: Path of the group (slug)
- `-d, --description string`: Description of the group
- `--visibility string`: Visibility level (public, internal, private)

### delete

Delete an existing group.

```bash
ash group delete <group-id-or-path>
```

### get

Get details of a specific group.

```bash
ash group get <group-id-or-path>
```

### clone

Clone a group and all its repositories.

```bash
ash group clone <group-id-or-path>
```

### sync

Sync all projects within a simple group or a list of groups defined in a file.

```bash
ash group sync
```
