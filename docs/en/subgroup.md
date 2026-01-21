# Subgroup Command

The `subgroup` command manages GitLab subgroups.

## Usage

```bash
ash subgroup [command]
```

## Available Commands

### list

List subgroups of a specific group.

```bash
ash subgroup list
```

### create

Create a new subgroup.

```bash
ash subgroup create <Name>
```

**Flags:**

- `--dir string`: Custom local directory name (default: same as subgroup name).
- `--visibility string`: Visibility level (public/internal/private) (default: `public`).

### delete

Delete a subgroup.

```bash
ash subgroup delete <Name or ID>
```

**Flags:**

- `-f, --force`: Force delete on GitLab.
- `-l, --local-force`: Also delete the local directory.

### clone

Clone a subgroup and all its repositories.

```bash
ash subgroup clone <Name or ID>
```

### sync

Sync all projects within a subgroup.

```bash
ash subgroup sync
```

**Flags:**

- `--clean`: Delete local folders of projects that identify as orphans.
