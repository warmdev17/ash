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
ash group create <Name>
```

*(No flags)*

### delete

Delete an existing group.

```bash
ash group delete <group-name (folder name), id or path>
```

**Flags:**

- `-f, --force`: Force delete on GitLab (even if not empty).
- `-l, --local-force`: Also delete the local directory.

### get

Get all managed groups information and save to config file.

```bash
ash group get
```

### clone

Clone a group and all its repositories.

```bash
ash group clone <Name or ID>
```

**Flags:**

- `--git-proto string`: Clone protocol (ssh/https) (default: `https`).

### sync

Sync all projects within a simple group or a list of groups defined in the config file.

```bash
ash group sync
```

**Flags:**

- `--clean`: Delete local folders of subgroups that identify as orphans (removed from GitLab).
