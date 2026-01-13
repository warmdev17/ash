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
ash subgroup list --group <parent-group-id>
```

### create

Create a new subgroup.

```bash
ash subgroup create
```

**Flags:**
- `-n, --name string`: Name of the subgroup
- `-p, --parent string`: ID or path of the parent group
- `-s, --slug string`: Path/Slug for the subgroup

### delete

Delete a subgroup.

```bash
ash subgroup delete <subgroup-id-or-path>
```

### clone

Clone a subgroup and all its repositories.

```bash
ash subgroup clone <subgroup-id-or-path>
```

### sync

Sync all projects within a subgroup.

```bash
ash subgroup sync <subgroup-id-or-path>
```
